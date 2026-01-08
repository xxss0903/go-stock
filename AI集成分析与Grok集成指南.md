# AI集成分析与Grok集成指南

## 📊 项目AI集成架构分析

### 1. AI集成方式

项目采用**两种方式**集成AI：

#### 方式一：直接API调用（OpenAI兼容格式）
- **位置**: `backend/data/openai_api.go`
- **核心函数**: 
  - `AskAi()` - 基础AI调用
  - `AskAiWithTools()` - 带工具函数的AI调用
  - `NewChatStream()` - 流式对话
- **特点**: 
  - 支持OpenAI兼容的API格式
  - 支持流式响应（Stream）
  - 支持工具函数调用（Function Calling）
  - 支持思考模式（Thinking Mode）

#### 方式二：Agent模式（使用cloudwego/eino框架）
- **位置**: `backend/agent/agent.go`
- **核心函数**: `GetStockAiAgent()`
- **特点**:
  - 使用ReAct Agent模式
  - 自动工具调用和决策
  - 支持多轮对话和工具链调用

### 2. AI配置结构

```go
type AIConfig struct {
    ID          uint    // 配置ID
    Name        string  // 配置名称
    BaseUrl     string  // API基础URL
    ApiKey      string  // API密钥
    ModelName   string  // 模型名称
    MaxTokens   int     // 最大token数
    Temperature float64 // 温度参数
    TimeOut     int     // 超时时间（秒）
}
```

### 3. AI工具函数（Tools）

项目提供了丰富的工具函数，AI可以调用这些工具获取数据：

1. **SearchStockByIndicators** - 根据技术指标筛选股票
2. **GetStockKLine** - 获取K线数据
3. **InteractiveAnswer** - 获取投资者互动数据
4. **GetStockResearchReport** - 获取股票研究报告
5. **HotStrategyTable** - 获取热门选股策略
6. **HotStockTable** - 获取热门股票排名

### 4. AI工作流程

```
用户提问
    ↓
收集股票数据（K线、财报、新闻等）
    ↓
构建消息上下文（System Prompt + 数据 + 用户问题）
    ↓
调用AI API（支持工具函数）
    ↓
AI分析并可能调用工具获取更多数据
    ↓
返回分析结果（流式输出）
    ↓
保存分析结果到数据库
```

---

## 🚀 集成Grok AI步骤

### 步骤1：了解Grok API

根据xAI官方信息，Grok API应该支持OpenAI兼容的格式。如果支持，可以直接使用现有代码。

**Grok API信息**（需要确认）：
- Base URL: `https://api.x.ai/v1` (需要确认)
- 模型名称: `grok-beta`, `grok-2`, `grok-3` 等
- API Key: 从xAI平台获取

### 步骤2：添加Grok配置

#### 方法A：如果Grok支持OpenAI兼容API（推荐）

**无需修改代码**，直接在应用设置中添加AI配置：

1. 打开应用设置
2. 添加新的AI配置：
   - **名称**: Grok AI
   - **Base URL**: `https://api.x.ai/v1` (或xAI提供的实际URL)
   - **API Key**: 你的Grok API密钥
   - **模型名称**: `grok-beta` 或 `grok-2` (根据可用模型)
   - **Max Tokens**: 4096 (根据模型限制调整)
   - **Temperature**: 0.7 (推荐值)
   - **Timeout**: 300

#### 方法B：如果需要特殊处理（如果Grok API格式不同）

需要修改 `backend/data/openai_api.go` 中的 `AskAi` 函数，添加Grok特定的处理逻辑。

### 步骤3：测试Grok集成

1. 在设置中选择Grok配置
2. 尝试分析一只股票
3. 检查日志文件 `logs/info.log` 查看是否有错误

---

## 💡 实现买卖策略功能

### 方案1：通过Prompt模板实现（最简单）

#### 1.1 创建策略分析Prompt模板

在应用设置中，创建或修改Prompt模板，添加策略分析指令：

```markdown
【角色设定】
你是一位专业的股票交易策略分析师，擅长技术分析、基本面分析和量化交易策略。

【任务要求】
根据提供的股票数据，给出明确的买卖策略建议，包括：

1. **买入信号**：
   - 买入理由（3-5点）
   - 建议买入价格区间
   - 建议买入仓位（占总资金比例）
   - 止损价位
   - 目标价位

2. **卖出信号**：
   - 卖出理由（3-5点）
   - 建议卖出价格区间
   - 建议卖出比例（全部/部分）
   - 止盈价位

3. **持仓建议**：
   - 当前是否适合持仓
   - 持仓周期建议（短线/中线/长线）
   - 风险等级评估

4. **策略执行计划**：
   - 具体操作步骤
   - 关键时间节点
   - 需要关注的风险点

【输出格式】
请使用以下格式输出：

## 📈 买入策略
- **买入理由**: [理由1] [理由2] [理由3]
- **买入价格**: XX.XX - XX.XX 元
- **建议仓位**: XX%
- **止损位**: XX.XX 元
- **目标位**: XX.XX 元

## 📉 卖出策略
- **卖出理由**: [理由1] [理由2] [理由3]
- **卖出价格**: XX.XX - XX.XX 元
- **卖出比例**: XX%
- **止盈位**: XX.XX 元

## ⚠️ 风险提示
[风险说明]

## 📅 执行计划
1. [步骤1]
2. [步骤2]
3. [步骤3]
```

#### 1.2 使用策略Prompt

在分析股票时，选择这个策略Prompt模板，AI会自动按照格式输出买卖策略。

### 方案2：添加策略分析工具函数（更强大）

#### 2.1 创建策略分析工具

在 `backend/agent/tools/` 目录下创建新文件 `trading_strategy_tool.go`：

```go
package tools

import (
    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/schema"
)

// GetTradingStrategyTool 获取交易策略工具
func GetTradingStrategyTool() tool.BaseTool {
    return tool.NewTool(
        tool.WithName("GetTradingStrategy"),
        tool.WithDescription("根据股票数据生成买卖策略，包括买入价格、卖出价格、止损止盈位、仓位建议等"),
        tool.WithInputSchema(&schema.ObjectSchema{
            Properties: map[string]schema.Schema{
                "stockCode": &schema.StringSchema{
                    Description: "股票代码",
                },
                "strategyType": &schema.StringSchema{
                    Description: "策略类型：short(短线)、medium(中线)、long(长线)",
                },
            },
            Required: []string{"stockCode"},
        }),
        tool.WithFunc(func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
            // 实现策略生成逻辑
            // 1. 获取股票数据
            // 2. 分析技术指标
            // 3. 生成策略建议
            // 4. 返回结构化策略数据
            return strategy, nil
        }),
    )
}
```

#### 2.2 在Agent中注册工具

修改 `backend/agent/agent.go`，添加策略工具：

```go
aiTools := compose.ToolsNodeConfig{
    Tools: []tool.BaseTool{
        // ... 现有工具
        tools.GetTradingStrategyTool(), // 添加策略工具
    },
}
```

#### 2.3 在app.go中注册工具

修改 `app.go` 中的 `AddTools` 函数，添加策略工具定义：

```go
tools = append(tools, data.Tool{
    Type: "function",
    Function: data.ToolFunction{
        Name:        "GetTradingStrategy",
        Description: "根据股票代码生成详细的买卖策略，包括买入卖出价格、止损止盈位、仓位建议等",
        Parameters: &data.FunctionParameters{
            Type: "object",
            Properties: map[string]any{
                "stockCode": map[string]any{
                    "type":        "string",
                    "description": "股票代码",
                },
                "strategyType": map[string]any{
                    "type":        "string",
                    "description": "策略类型：short(短线)、medium(中线)、long(长线)",
                },
            },
            Required: []string{"stockCode"},
        },
    },
})
```

### 方案3：创建专门的策略分析函数（最灵活）

#### 3.1 创建策略分析API

在 `backend/data/` 目录下创建 `trading_strategy_api.go`：

```go
package data

import (
    "context"
    "fmt"
    "go-stock/backend/logger"
)

// TradingStrategy 交易策略结构
type TradingStrategy struct {
    StockCode      string  `json:"stockCode"`
    StockName      string  `json:"stockName"`
    StrategyType   string  `json:"strategyType"` // short, medium, long
    BuySignal      bool    `json:"buySignal"`
    SellSignal     bool    `json:"sellSignal"`
    BuyPrice       float64 `json:"buyPrice"`
    SellPrice      float64 `json:"sellPrice"`
    StopLoss       float64 `json:"stopLoss"`
    TakeProfit     float64 `json:"takeProfit"`
    PositionSize   float64 `json:"positionSize"` // 仓位比例 0-1
    RiskLevel      string  `json:"riskLevel"`    // low, medium, high
    Reasoning      string  `json:"reasoning"`
    ExecutionPlan  string  `json:"executionPlan"`
}

// GenerateTradingStrategy 生成交易策略
func GenerateTradingStrategy(ctx context.Context, stockCode, stockName string, aiConfigId int) (*TradingStrategy, error) {
    // 1. 获取股票数据
    stockData, err := NewStockDataApi().GetStockCodeRealTimeData(stockCode)
    if err != nil {
        return nil, err
    }
    
    // 2. 获取K线数据
    klineData := NewStockDataApi().GetKLineData(stockCode, "240", 60)
    
    // 3. 获取财报数据
    // ...
    
    // 4. 构建策略分析Prompt
    prompt := fmt.Sprintf(`
请分析股票 %s[%s]，给出详细的买卖策略。

当前价格: %s
K线数据: [已提供]
财报数据: [已提供]

请按照以下格式输出策略：
1. 买入信号（是/否）及理由
2. 卖出信号（是/否）及理由
3. 建议买入价格区间
4. 建议卖出价格区间
5. 止损价位
6. 止盈价位
7. 建议仓位（0-100%）
8. 风险等级（低/中/高）
9. 执行计划

请以JSON格式输出。
`, stockName, stockCode, (*stockData)[0].Price)
    
    // 5. 调用AI生成策略
    ai := NewDeepSeekOpenAi(ctx, aiConfigId)
    // ... 调用AI并解析结果
    
    // 6. 返回策略对象
    return strategy, nil
}
```

#### 3.2 在前端添加策略显示

修改前端组件，添加策略显示区域。

---

## 🎯 推荐实施方案

### 快速方案（推荐新手）

**使用Prompt模板**：
1. 创建策略分析Prompt模板
2. 在分析股票时选择该模板
3. AI会自动输出策略建议

**优点**：
- 无需修改代码
- 立即可以使用
- 灵活调整Prompt

**缺点**：
- 输出格式可能不统一
- 需要手动解析策略信息

### 完整方案（推荐进阶）

**添加策略工具函数**：
1. 创建 `GetTradingStrategy` 工具函数
2. AI可以主动调用该工具
3. 返回结构化的策略数据
4. 前端可以格式化显示

**优点**：
- 输出格式统一
- 可以程序化处理
- 更智能的决策

**缺点**：
- 需要修改代码
- 需要测试和调试

---

## 📝 具体实现步骤（使用Prompt模板）

### 步骤1：创建策略Prompt模板

1. 打开应用
2. 进入"设置" -> "提示词模板"
3. 点击"添加"
4. 填写：
   - **名称**: 股票买卖策略分析
   - **类型**: system
   - **内容**: 使用上面提供的策略Prompt模板

### 步骤2：配置Grok AI

1. 进入"设置" -> "AI配置"
2. 点击"添加"
3. 填写Grok配置信息
4. 保存

### 步骤3：使用策略分析

1. 选择要分析的股票
2. 点击"AI分析"
3. 选择：
   - AI模型：Grok AI
   - 提示词模板：股票买卖策略分析
4. 输入问题或使用默认模板
5. 查看AI生成的策略建议

---

## 🔧 代码修改示例（如果需要）

如果需要添加Grok特定的处理，可以修改 `backend/data/openai_api.go`：

```go
func AskAi(o *OpenAi, err error, messages []map[string]interface{}, ch chan map[string]any, question string, think bool) {
    client := resty.New()
    client.SetBaseURL(strutil.Trim(o.BaseUrl))
    client.SetHeader("Authorization", "Bearer "+o.ApiKey)
    client.SetHeader("Content-Type", "application/json")
    
    // 如果是Grok，可能需要特殊处理
    if strings.Contains(o.BaseUrl, "x.ai") || strings.Contains(o.Model, "grok") {
        // Grok特定的配置
        // 例如：不同的请求格式、不同的参数等
    }
    
    // ... 其余代码
}
```

---

## ⚠️ 注意事项

1. **API兼容性**: 确认Grok API是否完全兼容OpenAI格式
2. **模型限制**: 了解Grok模型的token限制和功能限制
3. **成本考虑**: Grok API可能有使用费用，注意成本控制
4. **策略风险**: AI生成的策略仅供参考，不构成投资建议
5. **数据准确性**: 确保输入给AI的数据准确可靠

---

## 📚 相关文件位置

- AI配置: `backend/data/settings_api.go`
- AI调用: `backend/data/openai_api.go`
- Agent模式: `backend/agent/agent.go`
- 工具函数: `backend/agent/tools/`
- 前端设置: `frontend/src/components/settings.vue`

---

## 🎓 学习资源

- OpenAI API文档: https://platform.openai.com/docs
- xAI Grok文档: https://x.ai/docs (需要查看官方文档)
- 项目README: 查看项目README了解更多功能

---

**提示**: 建议先使用Prompt模板方案快速实现，如果效果满意，再考虑添加专门的策略工具函数。

