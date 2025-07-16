# 简化翻译系统

## 概述

ThreadMirror 的翻译系统现已完全简化，提供自动化的多语言支持：

- **自动触发**：Thread 抓取完成后自动翻译到多种语言
- **Entity 保护**：只翻译纯文本，保留 @mentions、#hashtags、URLs 等实体
- **简单架构**：移除复杂的分段和状态管理

## 核心特性

### 1. 自动翻译
当 thread 抓取完成时，系统会自动翻译到以下语言：
- 中文 (zh)
- 英文 (en) 
- 日文 (ja)
- 韩文 (ko)
- 西班牙文 (es)
- 法文 (fr)

### 2. Entity 感知
系统会自动识别并保留推文中的实体：
- `@用户名` 提及
- `#话题标签` 
- `https://...` 链接
- 媒体文件

**示例转换**：
```
原文: "Check out this amazing AI tool @openai #AI https://example.com"
翻译: "看看这个惊人的AI工具 @openai #AI https://example.com"
```

### 3. 数据结构

#### TweetText
```json
{
  "tweet_id": "123456789",
  "displayable_text": "This is the pure text content",
  "translated_text": "这是纯文本内容"
}
```

#### Translation
```json
{
  "id": "uuid",
  "thread_id": "123456789", 
  "source_language": "en",
  "target_language": "zh",
  "tweet_texts": [TweetText],
  "created_at": "2024-01-01T00:00:00Z"
}
```

## API 端点

### 获取 Thread 翻译列表
```http
GET /api/v1/thread/{id}/translations
```

### 获取特定语言翻译
```http
GET /api/v1/thread/{id}/translate/{target_language}
```

### 获取翻译详情
```http
GET /api/v1/translation/{id}
```

### 获取支持的语言
```http
GET /api/v1/translation/languages
```

## 技术实现

### 1. 翻译工具 (`pkg/util/translation.go`)
- 使用 `Tweet.GetDisplayableText()` 提取纯文本
- 自动过滤 entities

### 2. 翻译服务 (`internal/service/translation_service.go`)
- 批量翻译（5条推文/批次）
- 自动语言检测
- LLM 集成

### 3. 自动触发 (`internal/task/queue/thread_scrape_handler.go`)
- Thread 抓取完成后自动启动翻译
- 并发处理多种目标语言

## 优势

1. **简单直观**：无复杂的状态管理和分段逻辑
2. **自动化**：用户无需手动创建翻译
3. **Entity 保护**：社交媒体实体保持原样
4. **高效处理**：批量翻译提高性能
5. **多语言支持**：一次抓取，多语言翻译

## 配置

系统支持的语言在 `internal/model/translation.go` 中定义：

```go
func GetSupportedLanguages() []LanguageConfig {
    return []LanguageConfig{
        {Code: "zh", Name: "中文", IsSupported: true},
        {Code: "en", Name: "English", IsSupported: true},
        // ... 更多语言
    }
}
```

## 使用示例

1. **抓取 Thread**：
   ```bash
   POST /api/v1/thread/scrape
   {"url": "https://x.com/user/status/123456789"}
   ```

2. **自动翻译**：系统自动处理，无需额外操作

3. **获取中文翻译**：
   ```bash
   GET /api/v1/thread/123456789/translate/zh
   ```

4. **查看所有翻译**：
   ```bash
   GET /api/v1/thread/123456789/translations
   ```

这个简化的翻译系统完全满足了 "Thread 就是 TweetText 数组，Entity 不翻译，只翻译 DisplayableText，翻译在 thread 抓取完成后自动进行" 的要求，同时移除了所有冗余和复杂的向后兼容代码。 