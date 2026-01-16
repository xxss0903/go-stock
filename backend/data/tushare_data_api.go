package data

import (
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/go-resty/resty/v2"
	"go-stock/backend/logger"
	"strings"
	"time"
)

// @Author spark
// @Date 2025/2/17 12:33
// @Desc
//-----------------------------------------------------------------------------------

type TushareApi struct {
	client *resty.Client
	config *SettingConfig
}

func NewTushareApi(config *SettingConfig) *TushareApi {
	return &TushareApi{
		client: resty.New(),
		config: config,
	}
}

// GetDaily tushare A股日线行情
func (receiver TushareApi) GetDaily(tsCode, startDate, endDate string, crawlTimeOut int64) string {
	//logger.SugaredLogger.Debugf("tushare daily request: ts_code=%s, start_date=%s, end_date=%s", tsCode, startDate, endDate)
	fields := "ts_code,trade_date,open,high,low,close,pre_close,change,pct_chg,vol,amount"
	resp := &TushareStockBasicResponse{}
	stockType := getStockType(tsCode)
	tsCodeNEW := getTsCode(tsCode)
	//logger.SugaredLogger.Debugf("tushare daily request: %s,tsCode:%s,tsCodeNEW:%s", stockType, tsCode, tsCodeNEW)
	_, err := receiver.client.SetTimeout(time.Duration(crawlTimeOut)*time.Second).R().
		SetHeader("content-type", "application/json").
		SetBody(&TushareRequest{
			ApiName: stockType,
			Token:   receiver.config.TushareToken,
			Params: map[string]any{
				"ts_code":    tsCodeNEW,
				"start_date": startDate,
				"end_date":   endDate,
			},
			Fields: fields}).
		SetResult(resp).
		Post(tushareApiUrl)
	if err != nil {
		logger.SugaredLogger.Error(err)
		return ""
	}
	res := ""
	if resp.Data.Items != nil && len(resp.Data.Items) > 0 {
		fieldsStr := slice.JoinFunc(resp.Data.Fields, ",", func(s string) string {
			return "\"" + convertor.ToString(s) + "\""
		})
		res += fieldsStr + "\n"
		for _, item := range resp.Data.Items {
			//logger.SugaredLogger.Debugf("%s", slice.Join(item, ","))
			t := slice.JoinFunc(item, ",", func(s any) any {
				return "\"" + convertor.ToString(s) + "\""
			})
			res += t + "\n"
		}
	}
	//logger.SugaredLogger.Debugf("tushare response: %s", res)
	return res
}

func getTsCode(code string) any {
	if strutil.HasPrefixAny(code, []string{"US", "us", "gb_"}) {
		code = strings.Replace(code, "gb_", "", 1)
		code = strings.Replace(code, "us", "", 1)
		return code
	}
	return code
}

func getStockType(code string) string {
	if strutil.HasSuffixAny(code, []string{"SZ", "SH", "sh", "sz"}) {
		return "daily"
	}
	if strutil.HasSuffixAny(code, []string{"HK", "hk"}) {
		return "hk_daily"
	}
	if strutil.HasPrefixAny(code, []string{"US", "us", "gb_"}) {
		return "us_daily"
	}
	return ""
}

// GetAStockMarketTurnover 获取A股市场总成交额（通过上证指数和深证成指的成交额相加）
func (receiver TushareApi) GetAStockMarketTurnover(tradeDate string, crawlTimeOut int64) float64 {
	if tradeDate == "" {
		tradeDate = time.Now().Format("20060102")
	}
	
	// 获取上证指数和深证成指的成交额
	shAmount := receiver.getIndexDailyAmount("000001.SH", tradeDate, crawlTimeOut)
	szAmount := receiver.getIndexDailyAmount("399001.SZ", tradeDate, crawlTimeOut)
	
	totalAmount := shAmount + szAmount
	logger.SugaredLogger.Infof("A股总成交额: 上证指数成交额=%.2f, 深证成指成交额=%.2f, 总计=%.2f", shAmount, szAmount, totalAmount)
	
	return totalAmount
}

// getIndexDailyAmount 获取指数的日线成交额
func (receiver TushareApi) getIndexDailyAmount(indexCode, tradeDate string, crawlTimeOut int64) float64 {
	fields := "ts_code,trade_date,close,pct_chg,open,high,low,pre_close,change,vol,amount"
	resp := &TushareStockBasicResponse{}
	
	_, err := receiver.client.SetTimeout(time.Duration(crawlTimeOut)*time.Second).R().
		SetHeader("content-type", "application/json").
		SetBody(&TushareRequest{
			ApiName: "index_daily",
			Token:   receiver.config.TushareToken,
			Params: map[string]any{
				"ts_code":    indexCode,
				"trade_date": tradeDate,
			},
			Fields: fields}).
		SetResult(resp).
		Post(tushareApiUrl)
	
	if err != nil {
		logger.SugaredLogger.Errorf("获取指数成交额失败 %s: %s", indexCode, err.Error())
		return 0
	}
	
	if resp.Code != 0 {
		logger.SugaredLogger.Errorf("Tushare API返回错误 %s: %s", indexCode, resp.Msg)
		return 0
	}
	
	// 解析数据，amount字段在索引10
	if resp.Data.Items != nil && len(resp.Data.Items) > 0 {
		for _, item := range resp.Data.Items {
			if len(item) > 10 {
				amount, _ := convertor.ToFloat(item[10]) // amount字段
				return amount
			}
		}
	}
	
	return 0
}