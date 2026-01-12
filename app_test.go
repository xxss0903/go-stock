package main

import (
	"context"
	"encoding/json"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
)

// @Author spark
// @Date 2025/2/24 9:35
// @Desc
// -----------------------------------------------------------------------------------

func TestCheckStockBaseInfo(t *testing.T) {
	db.Init("./data/stock.db")
	NewApp().CheckStockBaseInfo(context.Background())
}


func TestUpdateCheck(t *testing.T) {
	releaseVersion := &models.GitHubReleaseVersion{}
	_, err := resty.New().R().
		SetResult(releaseVersion).
		SetHeader("Accept", "application/vnd.github+json").
		SetHeader("X-GitHub-Api-Version", "2022-11-28").
		Get("https://api.github.com/repos/ArvinLovegood/go-stock/releases/latest")
	//  https://api.github.com/repos/OWNER/REPO/releases/latest
	if err != nil {
		logger.SugaredLogger.Errorf("get github release version error:%s", err.Error())
		return
	}
	logger.SugaredLogger.Infof("releaseVersion:%+v", releaseVersion)
}
