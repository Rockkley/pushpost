package usecase

import (
	"context"
	"fmt"
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

func (u *ProfileUseCase) UploadAvatar(
	ctx context.Context,
	userID uuid.UUID,
	r io.Reader,
	size int64,
	contentType string,
) (string, error) {
	log := ctxlog.From(ctx).With(
		slog.String("op", "ProfileUseCase.UploadAvatar"),
		slog.String("user_id", userID.String()),
	)

	ext, ok := allowedMimeTypes[contentType]
	if !ok {
		return "", commonapperr.Validation(
			commonapperr.CodeFieldInvalid,
			"avatar",
			"unsupported file type; allowed: jpeg, png, webp",
		)
	}

	// Получаем текущий профиль, чтобы удалить старый аватар после загрузки.
	existing, err := u.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf("avatars/%s/%s%s", userID, uuid.New(), ext)

	avatarURL, err := u.storage.Upload(ctx, key, r, size, contentType)
	if err != nil {
		log.Error("failed to upload avatar to object storage", slog.Any("error", err))
		return "", commonapperr.Service("failed to upload avatar", err)
	}

	if err = u.profileRepo.UpdateAvatar(ctx, userID, avatarURL); err != nil {
		// Загрузка прошла, но БД не обновилась — удаляем осиротевший объект.
		if delErr := u.storage.Delete(ctx, key); delErr != nil {
			log.Error("failed to rollback orphaned avatar",
				slog.String("key", key),
				slog.Any("error", delErr),
			)
		}
		return "", err
	}

	// Удаляем старый аватар best-effort: ошибка не критична.
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

	log.Info("avatar uploaded", slog.String("url", avatarURL))
	return avatarURL, nil
}
