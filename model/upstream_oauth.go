package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	UpstreamOAuthSessionPending   = "pending"
	UpstreamOAuthSessionCompleted = "completed"
	UpstreamOAuthSessionFailed    = "failed"
	UpstreamOAuthSessionCancelled = "cancelled"
	UpstreamOAuthSessionExpired   = "expired"

	UpstreamCredentialEnabled  = 1
	UpstreamCredentialDisabled = 2
)

type UpstreamOAuthSession struct {
	Id                string `json:"id" gorm:"type:varchar(64);primaryKey"`
	ChannelId         int    `json:"channel_id" gorm:"index;not null"`
	AdminUserId       int    `json:"-" gorm:"index;not null"`
	Provider          string `json:"provider" gorm:"type:varchar(32);index;not null"`
	FlowType          string `json:"flow_type" gorm:"type:varchar(16);not null"`
	StateHash         string `json:"-" gorm:"type:varchar(64);uniqueIndex;not null"`
	EncryptedVerifier string `json:"-" gorm:"type:text"`
	EncryptedMetadata string `json:"-" gorm:"type:text"`
	Status            string `json:"status" gorm:"type:varchar(16);index;not null"`
	ErrorMessage      string `json:"error_message,omitempty" gorm:"type:varchar(512)"`
	CreatedAt         int64  `json:"created_at" gorm:"bigint;not null"`
	ExpiresAt         int64  `json:"expires_at" gorm:"bigint;index;not null"`
	CompletedAt       int64  `json:"completed_at,omitempty" gorm:"bigint"`
}

type UpstreamCredential struct {
	Id               int64  `json:"id"`
	ChannelId        int    `json:"channel_id" gorm:"uniqueIndex:idx_upstream_account,priority:1;index;not null"`
	Provider         string `json:"provider" gorm:"type:varchar(32);uniqueIndex:idx_upstream_account,priority:2;index;not null"`
	AccountId        string `json:"account_id" gorm:"type:varchar(255);uniqueIndex:idx_upstream_account,priority:3;not null"`
	AccountEmail     string `json:"account_email,omitempty" gorm:"type:varchar(255)"`
	DisplayName      string `json:"display_name,omitempty" gorm:"type:varchar(255)"`
	EncryptedPayload string `json:"-" gorm:"type:text;not null"`
	Status           int    `json:"status" gorm:"index;not null"`
	DisabledReason   string `json:"disabled_reason,omitempty" gorm:"type:varchar(512)"`
	ExpiresAt        int64  `json:"expires_at,omitempty" gorm:"bigint;index"`
	CooldownUntil    int64  `json:"cooldown_until,omitempty" gorm:"bigint;index"`
	RefreshingUntil  int64  `json:"-" gorm:"bigint;index"`
	LastSelectedAt   int64  `json:"last_selected_at,omitempty" gorm:"bigint;index"`
	LastSuccessAt    int64  `json:"last_success_at,omitempty" gorm:"bigint"`
	LastFailureAt    int64  `json:"last_failure_at,omitempty" gorm:"bigint"`
	FailureCount     int    `json:"failure_count"`
	QuotaMetadata    string `json:"quota_metadata,omitempty" gorm:"type:text"`
	ModelMetadata    string `json:"model_metadata,omitempty" gorm:"type:text"`
	Version          int64  `json:"version" gorm:"bigint;not null"`
	CreatedAt        int64  `json:"created_at" gorm:"bigint;not null"`
	UpdatedAt        int64  `json:"updated_at" gorm:"bigint;not null"`
}

type UpstreamCredentialModelState struct {
	Id            int64  `json:"id"`
	CredentialId  int64  `json:"credential_id" gorm:"uniqueIndex:idx_upstream_credential_model,priority:1;index;not null"`
	Model         string `json:"model" gorm:"type:varchar(191);uniqueIndex:idx_upstream_credential_model,priority:2;not null"`
	CooldownUntil int64  `json:"cooldown_until" gorm:"bigint;index"`
	Reason        string `json:"reason,omitempty" gorm:"type:varchar(512)"`
	UpdatedAt     int64  `json:"updated_at" gorm:"bigint;not null"`
}

func CreateUpstreamOAuthSession(session *UpstreamOAuthSession) error {
	return DB.Create(session).Error
}

func GetUpstreamOAuthSession(id string, adminUserId int) (*UpstreamOAuthSession, error) {
	var session UpstreamOAuthSession
	err := DB.Where("id = ? AND admin_user_id = ?", id, adminUserId).First(&session).Error
	return &session, err
}

func GetUpstreamOAuthSessionByStateHash(stateHash string) (*UpstreamOAuthSession, error) {
	var session UpstreamOAuthSession
	err := DB.Where("state_hash = ?", stateHash).First(&session).Error
	return &session, err
}

func UpdateUpstreamOAuthSession(id string, status string, errorMessage string) error {
	updates := map[string]any{"status": status, "error_message": errorMessage}
	if status != UpstreamOAuthSessionPending {
		updates["completed_at"] = time.Now().Unix()
	}
	return DB.Model(&UpstreamOAuthSession{}).Where("id = ? AND status = ?", id, UpstreamOAuthSessionPending).Updates(updates).Error
}

func ListUpstreamCredentials(channelId int) ([]*UpstreamCredential, error) {
	credentials := make([]*UpstreamCredential, 0)
	err := DB.Where("channel_id = ?", channelId).Order("id asc").Find(&credentials).Error
	return credentials, err
}

func GetUpstreamCredential(id int64, channelId int) (*UpstreamCredential, error) {
	var credential UpstreamCredential
	err := DB.Where("id = ? AND channel_id = ?", id, channelId).First(&credential).Error
	return &credential, err
}

func UpsertUpstreamCredential(credential *UpstreamCredential) error {
	return UpsertUpstreamCredentials([]*UpstreamCredential{credential})
}

func UpsertUpstreamCredentials(credentials []*UpstreamCredential) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		for _, credential := range credentials {
			if credential == nil {
				continue
			}
			if err := upsertUpstreamCredential(tx, credential); err != nil {
				return err
			}
		}
		return nil
	})
}

func upsertUpstreamCredential(tx *gorm.DB, credential *UpstreamCredential) error {
	var existing UpstreamCredential
	err := lockForUpdate(tx).Where("channel_id = ? AND provider = ? AND account_id = ?", credential.ChannelId, credential.Provider, credential.AccountId).First(&existing).Error
	now := time.Now().Unix()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		credential.Status = UpstreamCredentialEnabled
		credential.Version = 1
		credential.CreatedAt = now
		credential.UpdatedAt = now
		return tx.Create(credential).Error
	}
	if err != nil {
		return err
	}
	updates := map[string]any{
		"account_email": credential.AccountEmail, "display_name": credential.DisplayName,
		"encrypted_payload": credential.EncryptedPayload, "expires_at": credential.ExpiresAt,
		"status": UpstreamCredentialEnabled, "disabled_reason": "", "cooldown_until": 0,
		"failure_count": 0, "version": existing.Version + 1, "updated_at": now,
		"refreshing_until": 0,
	}
	credential.Id = existing.Id
	return tx.Model(&UpstreamCredential{}).Where("id = ?", existing.Id).Updates(updates).Error
}

func ClaimUpstreamCredentialRefresh(id int64, version int64, now int64, refreshingUntil int64) (bool, error) {
	result := DB.Model(&UpstreamCredential{}).
		Where("id = ? AND version = ? AND (refreshing_until = 0 OR refreshing_until <= ?)", id, version, now).
		Updates(map[string]any{
			"refreshing_until": refreshingUntil,
			"version":          gorm.Expr("version + ?", 1),
			"updated_at":       now,
		})
	return result.RowsAffected == 1, result.Error
}

func ReleaseUpstreamCredentialRefresh(id int64) error {
	return DB.Model(&UpstreamCredential{}).Where("id = ?", id).Update("refreshing_until", 0).Error
}

func SelectUpstreamCredential(channelId int, provider string, modelName string, now int64) (*UpstreamCredential, bool, error) {
	var total int64
	if err := DB.Model(&UpstreamCredential{}).Where("channel_id = ? AND provider = ?", channelId, provider).Count(&total).Error; err != nil {
		return nil, false, err
	}
	if total == 0 {
		return nil, false, nil
	}
	var credentials []*UpstreamCredential
	err := DB.Where("channel_id = ? AND provider = ? AND status = ? AND (cooldown_until = 0 OR cooldown_until <= ?)", channelId, provider, UpstreamCredentialEnabled, now).
		Order("last_selected_at asc, id asc").Find(&credentials).Error
	if err != nil {
		return nil, true, err
	}
	for _, credential := range credentials {
		if credential == nil {
			continue
		}
		if modelName != "" {
			var blocked int64
			if err = DB.Model(&UpstreamCredentialModelState{}).
				Where("credential_id = ? AND model = ? AND cooldown_until > ?", credential.Id, modelName, now).
				Count(&blocked).Error; err != nil {
				return nil, true, err
			}
			if blocked > 0 {
				continue
			}
		}
		if err = DB.Model(&UpstreamCredential{}).Where("id = ?", credential.Id).Updates(map[string]any{"last_selected_at": now, "updated_at": now}).Error; err != nil {
			return nil, true, err
		}
		return credential, true, nil
	}
	return nil, true, errors.New("no available OAuth credential")
}

func UpdateUpstreamCredentialStatus(id int64, channelId int, status int, reason string) error {
	return DB.Model(&UpstreamCredential{}).Where("id = ? AND channel_id = ?", id, channelId).Updates(map[string]any{
		"status": status, "disabled_reason": reason, "updated_at": time.Now().Unix(),
	}).Error
}

func DeleteUpstreamCredential(id int64, channelId int) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ? AND channel_id = ?", id, channelId).Delete(&UpstreamCredential{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Where("credential_id = ?", id).Delete(&UpstreamCredentialModelState{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func deleteUpstreamOAuthForChannels(tx *gorm.DB, channelIds []int) error {
	if len(channelIds) == 0 || !tx.Migrator().HasTable(&UpstreamCredential{}) {
		return nil
	}
	var credentialIds []int64
	if err := tx.Model(&UpstreamCredential{}).Where("channel_id IN ?", channelIds).Pluck("id", &credentialIds).Error; err != nil {
		return err
	}
	if len(credentialIds) > 0 && tx.Migrator().HasTable(&UpstreamCredentialModelState{}) {
		if err := tx.Where("credential_id IN ?", credentialIds).Delete(&UpstreamCredentialModelState{}).Error; err != nil {
			return err
		}
	}
	if err := tx.Where("channel_id IN ?", channelIds).Delete(&UpstreamCredential{}).Error; err != nil {
		return err
	}
	if tx.Migrator().HasTable(&UpstreamOAuthSession{}) {
		return tx.Where("channel_id IN ?", channelIds).Delete(&UpstreamOAuthSession{}).Error
	}
	return nil
}

func MarkUpstreamCredentialFailure(id int64, reason string, cooldownUntil int64) error {
	if len(reason) > 500 {
		reason = reason[:500]
	}
	now := time.Now().Unix()
	return DB.Model(&UpstreamCredential{}).Where("id = ?", id).Updates(map[string]any{
		"failure_count": gorm.Expr("failure_count + ?", 1), "last_failure_at": now,
		"cooldown_until": cooldownUntil, "disabled_reason": reason, "updated_at": now,
	}).Error
}

func MarkUpstreamCredentialSuccess(id int64) error {
	now := time.Now().Unix()
	return DB.Model(&UpstreamCredential{}).Where("id = ?", id).Updates(map[string]any{
		"failure_count": 0, "last_success_at": now, "cooldown_until": 0,
		"disabled_reason": "", "updated_at": now,
	}).Error
}

func MarkUpstreamCredentialModelFailure(credentialId int64, modelName string, reason string, cooldownUntil int64) error {
	if len(reason) > 500 {
		reason = reason[:500]
	}
	state := &UpstreamCredentialModelState{
		CredentialId: credentialId, Model: modelName, CooldownUntil: cooldownUntil,
		Reason: reason, UpdatedAt: time.Now().Unix(),
	}
	return DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "credential_id"}, {Name: "model"}},
		DoUpdates: clause.AssignmentColumns([]string{"cooldown_until", "reason", "updated_at"}),
	}).Create(state).Error
}

func ClearUpstreamCredentialModelFailure(credentialId int64, modelName string) error {
	if modelName == "" {
		return nil
	}
	return DB.Where("credential_id = ? AND model = ?", credentialId, modelName).Delete(&UpstreamCredentialModelState{}).Error
}
