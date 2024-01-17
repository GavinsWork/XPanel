package job

import (
	"encoding/json"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cloud_storage/client"
)

type onedrive struct {
}

func NewOneDriveJob() *onedrive {
	return &onedrive{}
}

func (onedrive *onedrive) Run() {
	var backupItem model.BackupAccount
	_ = global.DB.Where("`type` = ?", "OneDrive").First(&backupItem)
	if backupItem.ID == 0 {
		return
	}
	if len(backupItem.Credential) == 0 {
		global.LOG.Error("OneDrive configuration lacks token information, please rebind.")
		return
	}
	global.LOG.Info("start to refresh token of OneDrive ...")
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(backupItem.Vars), &varMap); err != nil {
		global.LOG.Errorf("Failed to refresh OneDrive token, please retry, err: %v", err)
		return
	}
	refreshItem, ok := varMap["refresh_token"]
	if !ok {
		global.LOG.Error("Failed to refresh OneDrive token, please retry, err: no such refresh token")
		return
	}

	token, refreshToken, err := client.RefreshToken("refresh_token", refreshItem.(string))
	varMap["refresh_status"] = constant.StatusSuccess
	varMap["refresh_time"] = time.Now().Format("2006-01-02 15:04:05")
	if err != nil {
		varMap["refresh_status"] = constant.StatusFailed
		varMap["refresh_msg"] = err.Error()
		global.LOG.Errorf("Failed to refresh OneDrive token, please retry, err: %v", err)
		return
	}
	varMap["refresh_token"] = refreshToken

	varsItem, _ := json.Marshal(varMap)
	_ = global.DB.Model(&model.Group{}).
		Where("`type` = ?", "OneDrive").
		Updates(map[string]interface{}{
			"credential": token,
			"vars":       varsItem,
		}).Error
	global.LOG.Info("Successfully refreshed OneDrive token.")
}
