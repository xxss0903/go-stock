package data

import (
	"encoding/json"
	"fmt"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	"time"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/go-resty/resty/v2"
	"github.com/robertkrimen/otto"
)

// @Author spark
// @Date 2025/1/8
// @Desc 炒股复盘记录API
//-----------------------------------------------------------------------------------

type TradingRecordApi struct {
	client *resty.Client
	config *SettingConfig
}

func NewTradingRecordApi() *TradingRecordApi {
	return &TradingRecordApi{
		client: resty.New(),
		config: GetSettingConfig(),
	}
}

// GetLimitUpDownSectors 获取涨停跌停板块信息
func (t *TradingRecordApi) GetLimitUpDownSectors(tradeDate string) (limitUpSectors []models.LimitUpDownSector, limitDownSectors []models.LimitUpDownSector, err error) {
	if tradeDate == "" {
		tradeDate = time.Now().Format("2006-01-02")
	}

	// 从东方财富获取涨停跌停数据
	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/clist/get?np=1&fltt=2&invt=2&cb=data&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23&fields=f12,f14,f2,f3,f62,f184,f66,f69,f72,f75,f78,f81,f84,f87,f100,f265&fid=f3&po=1&pz=5000&pn=1&_=%d", time.Now().UnixMilli())
	
	resp, err := t.client.SetTimeout(time.Duration(t.config.CrawlTimeOut)*time.Second).R().
		SetHeader("Host", "push2.eastmoney.com").
		SetHeader("Referer", "https://quote.eastmoney.com/center/gridlist.html").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:146.0) Gecko/20100101 Firefox/146.0").
		Get(url)
	
	if err != nil {
		logger.SugaredLogger.Errorf("获取涨停跌停数据失败: %s", err.Error())
		return nil, nil, err
	}

	body := string(resp.Body())
	// 使用otto解析JavaScript回调函数
	vm := otto.New()
	vm.Run("function data(res){return res};")
	val, err := vm.Run(body)
	if err != nil {
		logger.SugaredLogger.Errorf("解析JavaScript失败: %s", err.Error())
		return nil, nil, err
	}

	value, err := val.Object().Value().Export()
	if err != nil {
		logger.SugaredLogger.Errorf("导出数据失败: %s", err.Error())
		return nil, nil, err
	}

	marshal, err := json.Marshal(value)
	if err != nil {
		logger.SugaredLogger.Errorf("序列化数据失败: %s", err.Error())
		return nil, nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(marshal, &result); err != nil {
		logger.SugaredLogger.Errorf("解析涨停跌停数据失败: %s", err.Error())
		return nil, nil, err
	}

	// 解析数据
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		logger.SugaredLogger.Warnf("数据格式错误，result: %+v", result)
		return nil, nil, fmt.Errorf("数据格式错误")
	}

	diffs, ok := data["diff"].([]interface{})
	if !ok {
		logger.SugaredLogger.Warnf("diff数据格式错误，data: %+v", data)
		return nil, nil, fmt.Errorf("diff数据格式错误")
	}

	// 按板块分组统计涨停跌停
	sectorLimitUpMap := make(map[string][]string)
	sectorLimitDownMap := make(map[string][]string)

	for _, item := range diffs {
		stock := item.(map[string]interface{})
		code := convertor.ToString(stock["f12"])
		name := convertor.ToString(stock["f14"])
		pctChg, _ := convertor.ToFloat(stock["f3"]) // 涨跌幅
		bkName := convertor.ToString(stock["f100"]) // 板块名称

		if bkName == "" {
			continue
		}

		// 涨停（涨幅>=9.5%）
		if pctChg >= 9.5 {
			if _, ok := sectorLimitUpMap[bkName]; !ok {
				sectorLimitUpMap[bkName] = make([]string, 0)
			}
			sectorLimitUpMap[bkName] = append(sectorLimitUpMap[bkName], fmt.Sprintf("%s(%s)", name, code))
		}

		// 跌停（跌幅<=-9.5%）
		if pctChg <= -9.5 {
			if _, ok := sectorLimitDownMap[bkName]; !ok {
				sectorLimitDownMap[bkName] = make([]string, 0)
			}
			sectorLimitDownMap[bkName] = append(sectorLimitDownMap[bkName], fmt.Sprintf("%s(%s)", name, code))
		}
	}

	// 转换为结构体
	for sectorName, stocks := range sectorLimitUpMap {
		limitUpSectors = append(limitUpSectors, models.LimitUpDownSector{
			SectorName: sectorName,
			StockCount: len(stocks),
			Stocks:     stocks,
		})
	}

	for sectorName, stocks := range sectorLimitDownMap {
		limitDownSectors = append(limitDownSectors, models.LimitUpDownSector{
			SectorName: sectorName,
			StockCount: len(stocks),
			Stocks:     stocks,
		})
	}

	return limitUpSectors, limitDownSectors, nil
}

// CreateTradingRecord 创建复盘记录
func (t *TradingRecordApi) CreateTradingRecord(record *models.TradingRecord) error {
	if record.TradeDate == "" {
		record.TradeDate = time.Now().Format("2006-01-02")
	}
	return db.Dao.Model(&models.TradingRecord{}).Create(record).Error
}

// UpdateTradingRecord 更新复盘记录
func (t *TradingRecordApi) UpdateTradingRecord(record *models.TradingRecord) error {
	return db.Dao.Model(&models.TradingRecord{}).Where("id = ?", record.ID).Updates(record).Error
}

// DeleteTradingRecord 删除复盘记录
func (t *TradingRecordApi) DeleteTradingRecord(id uint) error {
	return db.Dao.Model(&models.TradingRecord{}).Where("id = ?", id).Delete(&models.TradingRecord{}).Error
}

// GetTradingRecord 获取单条复盘记录
func (t *TradingRecordApi) GetTradingRecord(id uint) (*models.TradingRecord, error) {
	var record models.TradingRecord
	err := db.Dao.Model(&models.TradingRecord{}).Where("id = ?", id).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetTradingRecordByDate 根据日期获取复盘记录
func (t *TradingRecordApi) GetTradingRecordByDate(tradeDate string) (*models.TradingRecord, error) {
	var record models.TradingRecord
	err := db.Dao.Model(&models.TradingRecord{}).Where("trade_date = ?", tradeDate).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetTradingRecordList 获取复盘记录列表
func (t *TradingRecordApi) GetTradingRecordList(page, pageSize int) ([]models.TradingRecord, int64, error) {
	var records []models.TradingRecord
	var total int64

	offset := (page - 1) * pageSize
	err := db.Dao.Model(&models.TradingRecord{}).
		Order("trade_date desc").
		Limit(pageSize).
		Offset(offset).
		Find(&records).Error
	
	if err != nil {
		return nil, 0, err
	}

	db.Dao.Model(&models.TradingRecord{}).Count(&total)
	return records, total, nil
}

