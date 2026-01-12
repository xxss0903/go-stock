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

// TushareLimitListResponse Tushare涨跌停列表响应
type TushareLimitListResponse struct {
	RequestId string `json:"request_id"`
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Data      struct {
		Fields []string   `json:"fields"`
		Items  [][]any    `json:"items"`
	} `json:"data"`
}

// LimitStockInfo 涨跌停股票信息
type LimitStockInfo struct {
	TradeDate     string  `json:"tradeDate"`     // 交易日期
	TsCode        string  `json:"tsCode"`        // 股票代码
	Industry      string  `json:"industry"`      // 所属行业
	Name          string  `json:"name"`          // 股票名称
	Close         float64 `json:"close"`         // 收盘价
	PctChg        float64 `json:"pctChg"`        // 涨跌幅
	Amount        float64 `json:"amount"`        // 成交额
	LimitAmount   float64 `json:"limitAmount"`  // 板上成交金额
	FloatMv       float64 `json:"floatMv"`       // 流通市值
	TotalMv       float64 `json:"totalMv"`       // 总市值
	TurnoverRatio float64 `json:"turnoverRatio"` // 换手率
	FdAmount      float64 `json:"fdAmount"`      // 封单金额
	FirstTime     string  `json:"firstTime"`     // 首次封板时间
	LastTime      string  `json:"lastTime"`      // 最后封板时间
	OpenTimes     int     `json:"openTimes"`     // 炸板次数
	UpStat        string  `json:"upStat"`        // 涨停统计
	LimitTimes    int     `json:"limitTimes"`    // 连板数
	LimitType     string  `json:"limitType"`     // D跌停U涨停Z炸板
}

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

	// 将日期格式转换为东方财富API需要的格式 (YYYYMMDD)
	dateTime, err := time.Parse("2006-01-02", tradeDate)
	if err != nil {
		logger.SugaredLogger.Errorf("日期格式错误: %s", err.Error())
		return nil, nil, fmt.Errorf("日期格式错误: %s", err.Error())
	}
	dateStr := dateTime.Format("20060102")

	// 从东方财富数据中心API获取指定日期的涨停跌停数据
	// 使用datacenter API获取历史数据
	url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?sortColumns=CHANGE_RATE&sortTypes=-1&pageSize=5000&pageNumber=1&reportName=RPT_DAILYBILLBOARD_DETAILSNEW&columns=SECURITY_CODE,SECUCODE,SECURITY_NAME_ABBR,TRADE_DATE,CLOSE_PRICE,CHANGE_RATE,BILLBOARD_NET_AMT,BILLBOARD_BUY_AMT,BILLBOARD_SELL_AMT&filter=(TRADE_DATE='%s')", dateStr)
	
	logger.SugaredLogger.Infof("获取涨停跌停数据，日期: %s, URL: %s", tradeDate, url)
	
	resp, err := t.client.SetTimeout(time.Duration(t.config.CrawlTimeOut)*time.Second).R().
		SetHeader("Host", "datacenter-web.eastmoney.com").
		SetHeader("Referer", "https://data.eastmoney.com/stock/tradedetail.html").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:146.0) Gecko/20100101 Firefox/146.0").
		Get(url)
	
	if err != nil {
		logger.SugaredLogger.Errorf("获取涨停跌停数据失败: %s", err.Error())
		// 如果历史数据API失败，且是当天数据，尝试使用实时API
		if dateStr == time.Now().Format("20060102") {
			return t.getCurrentDayLimitUpDownSectors()
		}
		return nil, nil, fmt.Errorf("获取历史数据失败: %s", err.Error())
	}

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		logger.SugaredLogger.Errorf("解析涨停跌停数据失败: %s", err.Error())
		// 如果解析失败，且是当天数据，尝试使用实时API
		if dateStr == time.Now().Format("20060102") {
			return t.getCurrentDayLimitUpDownSectors()
		}
		return nil, nil, fmt.Errorf("解析历史数据失败: %s", err.Error())
	}

	// 检查是否有数据
	successVal, ok := result["success"]
	if !ok {
		logger.SugaredLogger.Warnf("API返回数据缺少success字段，日期: %s", tradeDate)
		if dateStr == time.Now().Format("20060102") {
			return t.getCurrentDayLimitUpDownSectors()
		}
		return nil, nil, fmt.Errorf("该日期(%s)暂无数据", tradeDate)
	}
	// 将interface{}转换为字符串再转换为bool
	successStr := convertor.ToString(successVal)
	success, _ := convertor.ToBool(successStr)
	if !success {
		logger.SugaredLogger.Warnf("API返回失败，日期: %s", tradeDate)
		// 如果历史数据API没有数据，且是当天数据，尝试使用实时API
		if dateStr == time.Now().Format("20060102") {
			return t.getCurrentDayLimitUpDownSectors()
		}
		return nil, nil, fmt.Errorf("该日期(%s)暂无数据", tradeDate)
	}

	data, ok := result["result"].(map[string]interface{})
	if !ok {
		logger.SugaredLogger.Warnf("数据格式错误，尝试使用实时API")
		if dateStr == time.Now().Format("20060102") {
			return t.getCurrentDayLimitUpDownSectors()
		}
		return nil, nil, fmt.Errorf("数据格式错误")
	}

	stocks, ok := data["data"].([]interface{})
	if !ok {
		logger.SugaredLogger.Warnf("股票数据格式错误，尝试使用实时API")
		if dateStr == time.Now().Format("20060102") {
			return t.getCurrentDayLimitUpDownSectors()
		}
		return nil, nil, fmt.Errorf("股票数据格式错误")
	}

	// 获取所有股票的板块信息（需要单独查询）
	// 先获取涨停跌停股票列表
	sectorLimitUpMap := make(map[string][]string)
	sectorLimitDownMap := make(map[string][]string)

	for _, item := range stocks {
		stock := item.(map[string]interface{})
		code := convertor.ToString(stock["SECURITY_CODE"])
		name := convertor.ToString(stock["SECURITY_NAME_ABBR"])
		pctChg, _ := convertor.ToFloat(stock["CHANGE_RATE"])

		// 获取股票板块信息
		bkName := t.getStockSector(code)
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

// getCurrentDayLimitUpDownSectors 获取当天的涨停跌停数据（使用实时API）
func (t *TradingRecordApi) getCurrentDayLimitUpDownSectors() (limitUpSectors []models.LimitUpDownSector, limitDownSectors []models.LimitUpDownSector, err error) {
	// 从东方财富获取实时涨停跌停数据
	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/clist/get?np=1&fltt=2&invt=2&cb=data&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23&fields=f12,f14,f2,f3,f62,f184,f66,f69,f72,f75,f78,f81,f84,f87,f100,f265&fid=f3&po=1&pz=5000&pn=1&_=%d", time.Now().UnixMilli())
	
	resp, err := t.client.SetTimeout(time.Duration(t.config.CrawlTimeOut)*time.Second).R().
		SetHeader("Host", "push2.eastmoney.com").
		SetHeader("Referer", "https://quote.eastmoney.com/center/gridlist.html").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:146.0) Gecko/20100101 Firefox/146.0").
		Get(url)
	
	if err != nil {
		logger.SugaredLogger.Errorf("获取实时涨停跌停数据失败: %s", err.Error())
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

// getStockSector 获取股票所属板块
func (t *TradingRecordApi) getStockSector(stockCode string) string {
	// 从数据库查询股票板块信息
	var stockBasic StockBasic
	err := db.Dao.Model(&StockBasic{}).Where("symbol = ?", stockCode).First(&stockBasic).Error
	if err == nil && stockBasic.BKName != "" {
		return stockBasic.BKName
	}
	
	// 如果数据库没有，尝试从API获取（这里可以调用现有的获取股票信息的方法）
	// 暂时返回空，后续可以优化
	return ""
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
		Where("is_del = ?", 0).
		Order("trade_date desc").
		Limit(pageSize).
		Offset(offset).
		Find(&records).Error
	
	if err != nil {
		return nil, 0, err
	}

	db.Dao.Model(&models.TradingRecord{}).Where("is_del = ?", 0).Count(&total)
	return records, total, nil
}

// GetLimitListFromTushare 从Tushare获取涨跌停股票列表
func (t *TradingRecordApi) GetLimitListFromTushare(tradeDate string, limitType string) ([]LimitStockInfo, error) {
	if tradeDate == "" {
		tradeDate = time.Now().Format("20060102")
	} else {
		// 将日期格式从 2006-01-02 转换为 20060102
		dateTime, err := time.Parse("2006-01-02", tradeDate)
		if err != nil {
			logger.SugaredLogger.Errorf("日期格式错误: %s", err.Error())
			return nil, fmt.Errorf("日期格式错误: %s", err.Error())
		}
		tradeDate = dateTime.Format("20060102")
	}

	if limitType == "" {
		limitType = "U" // 默认获取涨停
	}

	// 添加延迟，避免API限流（Tushare限制每秒最多1次请求）
	time.Sleep(1100 * time.Millisecond)

	// 构建请求参数
	params := map[string]any{
		"trade_date": tradeDate,
		"limit_type": limitType,
	}

	fields := "ts_code,trade_date,industry,name,close,pct_chg,amount,limit_amount,float_mv,total_mv,turnover_ratio,fd_amount,first_time,last_time,open_times,up_stat,limit_times,limit"

	resp := &TushareLimitListResponse{}
	_, err := t.client.SetTimeout(time.Duration(t.config.CrawlTimeOut)*time.Second).R().
		SetHeader("content-type", "application/json").
		SetBody(map[string]any{
			"api_name": "limit_list_d",
			"token":    t.config.TushareToken,
			"params":   params,
			"fields":   fields,
		}).
		SetResult(resp).
		Post("http://api.tushare.pro")

	if err != nil {
		logger.SugaredLogger.Errorf("获取Tushare涨跌停数据失败: %s", err.Error())
		return nil, err
	}

	if resp.Code != 0 {
		logger.SugaredLogger.Errorf("Tushare API返回错误: %s", resp.Msg)
		return nil, fmt.Errorf("Tushare API错误: %s", resp.Msg)
	}

	// 解析数据
	var limitStocks []LimitStockInfo
	if resp.Data.Items != nil && len(resp.Data.Items) > 0 {
		for _, item := range resp.Data.Items {
			if len(item) < 18 {
				continue
			}

			close, _ := convertor.ToFloat(item[4])
			pctChg, _ := convertor.ToFloat(item[5])
			amount, _ := convertor.ToFloat(item[6])
			limitAmount, _ := convertor.ToFloat(item[7])
			floatMv, _ := convertor.ToFloat(item[8])
			totalMv, _ := convertor.ToFloat(item[9])
			turnoverRatio, _ := convertor.ToFloat(item[10])
			fdAmount, _ := convertor.ToFloat(item[11])
			openTimes, _ := convertor.ToInt(item[14])
			limitTimes, _ := convertor.ToInt(item[16])

			stock := LimitStockInfo{
				TsCode:        convertor.ToString(item[0]),
				TradeDate:     convertor.ToString(item[1]),
				Industry:      convertor.ToString(item[2]),
				Name:          convertor.ToString(item[3]),
				Close:         close,
				PctChg:        pctChg,
				Amount:        amount,
				LimitAmount:   limitAmount,
				FloatMv:       floatMv,
				TotalMv:       totalMv,
				TurnoverRatio: turnoverRatio,
				FdAmount:      fdAmount,
				FirstTime:     convertor.ToString(item[12]),
				LastTime:      convertor.ToString(item[13]),
				OpenTimes:     int(openTimes),
				UpStat:        convertor.ToString(item[15]),
				LimitTimes:    int(limitTimes),
				LimitType:     convertor.ToString(item[17]),
			}
			limitStocks = append(limitStocks, stock)
		}
	}

	logger.SugaredLogger.Infof("获取到 %d 条涨跌停数据", len(limitStocks))
	return limitStocks, nil
}

