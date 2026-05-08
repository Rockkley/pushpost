package usecase

import (
	"bytes"
	"context"
	"fmt"
	"github.com/rockkley/pushpost/services/profile_service/internal/domain/dto"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
	"image"
	"image/jpeg"
	"io"
	"log/slog"

	"github.com/google/uuid"
	commonapperr "github.com/rockkley/pushpost/services/common_service/apperror"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	"github.com/rockkley/pushpost/services/profile_service/internal/entity"
	"github.com/rockkley/pushpost/services/profile_service/internal/repository"
	"github.com/rockkley/pushpost/services/profile_service/internal/storage/minio"
)

var allowedMimeTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type ProfileUseCase struct {
	profileRepo repository.ProfileRepositoryInterface
	storage     minio.ObjectStorage
	keyFromURL  func(url string) string
}

func NewProfileUseCase(
	profileRepo repository.ProfileRepositoryInterface,
	objStorage minio.ObjectStorage,
	keyFromURL func(url string) string,
) *ProfileUseCase {
	return &ProfileUseCase{
		profileRepo: profileRepo,
		storage:     objStorage,
		keyFromURL:  keyFromURL,
	}
}

func (u *ProfileUseCase) GetByUsername(ctx context.Context, username string) (*entity.Profile, error) {
	return u.profileRepo.FindByUsername(ctx, username)
}

func (u *ProfileUseCase) CreateProfile(ctx context.Context, profile *entity.Profile) error {
	return u.profileRepo.Create(ctx, profile)
}

func (u *ProfileUseCase) UpdateProfile(ctx context.Context, profile *entity.Profile) error {
	return u.profileRepo.Update(ctx, profile)
}

func (u *ProfileUseCase) Search(ctx context.Context, filter *dto.SearchProfilesQuery) ([]*entity.Profile, error) {
	return u.profileRepo.Search(ctx, filter)
}

func (u *ProfileUseCase) UploadAvatar(
	ctx context.Context,
	userID uuid.UUID,
	r io.Reader,
	size int64,
	contentType string,
) (string, string, error) {
	log := ctxlog.From(ctx).With(
		slog.String("op", "ProfileUseCase.UploadAvatar"),
		slog.String("user_id", userID.String()),
	)

	ext, ok := allowedMimeTypes[contentType]
	if !ok {
		return "", "", commonapperr.Validation(
			commonapperr.CodeFieldInvalid,
			"avatar",
			"unsupported file type; allowed: jpeg, png, webp",
		)
	}

	existing, err := u.profileRepo.FindByUserID(ctx, userID)

	if err != nil {
		return "", "", err
	}

	original, err := io.ReadAll(r)

	if err != nil {
		return "", "", commonapperr.Service("failed to read avatar", err)
	}

	key := fmt.Sprintf("avatars/%s/%s%s", userID, uuid.New(), ext)
	thumbKey := fmt.Sprintf("avatars/%s/%s_thumb.jpg", userID, uuid.New())

	avatarURL, err := u.storage.Upload(ctx, key, bytes.NewReader(original), size, contentType)

	if err != nil {
		log.Error("failed to upload avatar to object storage", slog.Any("error", err))

		return "", "", commonapperr.Service("failed to upload avatar", err)
	}

	thumb, err := buildThumbnail(original)

	if err != nil {
		if delErr := u.storage.Delete(ctx, key); delErr != nil {
			log.Warn("failed to cleanup original avatar after thumb build failure", slog.Any("error", delErr))
		}
		return "", "", commonapperr.Validation(commonapperr.CodeFieldInvalid, "avatar", "failed to process avatar image")
	}

	avatarThumbURL, err := u.storage.Upload(ctx, thumbKey, bytes.NewReader(thumb), int64(len(thumb)), "image/jpeg")
	if err != nil {
		if delErr := u.storage.Delete(ctx, key); delErr != nil {
			log.Warn("failed to cleanup original avatar after thumb upload failure", slog.Any("error", delErr))
		}
		return "", "", commonapperr.Service("failed to upload avatar thumbnail", err)
	}

	if err = u.profileRepo.UpdateAvatar(ctx, userID, avatarURL, avatarThumbURL); err != nil {

		// Загрузка прошла, но БД не обновилась - удаляем осиротевший объект.
		if delErr := u.storage.Delete(ctx, key); delErr != nil {
			log.Error("failed to rollback orphaned avatar",
				slog.String("key", key),
				slog.Any("error", delErr),
			)
		}
		if delErr := u.storage.Delete(ctx, thumbKey); delErr != nil {
			log.Error("failed to rollback orphaned avatar thumbnail", slog.String("key", thumbKey), slog.Any("error", delErr))
		}
		return "", "", err

	}

	// Удаляем старый аватар best-effort: ошибка не критична. //fixme
	if existing.AvatarURL != nil {
		if oldKey := u.keyFromURL(*existing.AvatarURL); oldKey != "" {
			if delErr := u.storage.Delete(ctx, oldKey); delErr != nil {
				log.Warn("failed to delete old avatar",
					slog.String("old_key", oldKey),
					slog.Any("error", delErr),
				)
			}
		}
	}
	if existing.AvatarThumbURL != nil {
		if oldKey := u.keyFromURL(*existing.AvatarThumbURL); oldKey != "" {
			if delErr := u.storage.Delete(ctx, oldKey); delErr != nil {
				log.Warn("failed to delete old avatar thumbnail", slog.String("old_thumb_key", oldKey), slog.Any("error", delErr))
			}
		}
	}

	log.Info("avatar uploaded", slog.String("url", avatarURL))

	return avatarURL, avatarThumbURL, nil
}

func buildThumbnail(src []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}

	const thumbSize = 96
	dst := image.NewRGBA(image.Rect(0, 0, thumbSize, thumbSize))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	var out bytes.Buffer
	if err = jpeg.Encode(&out, dst, &jpeg.Options{Quality: 82}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
