# 中国节假日 api

方便使用和部署的中国节假日api

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https%3A%2F%2Fgithub.com%2Fpaulzhn%2Fchinese-holiday-api)

## 快速使用

推荐使用vercel进行一键部署，点击上方蓝色按钮即可。

注意：vercel默认域名在中国大陆可能无法正常访问，建议绑定自己的域名。

示例域名及请求（不保证可用性）：https://api.gointo.icu/api/holiday?date=2023-05-01


## 接口文档

URI: `/api/holiday`

请求方式：GET

参数：（均为query）

- `verbose`：信息的详细程度，取值为0 ~ 2，默认为 0
  - 0：仅返回是否为节假日（含周末），0 - 否，1 - 是
  - 1：返回日期类型的枚举值，枚举值取值范围为 0 ~ 4
    - 0：普通工作日
    - 1：普通周末
    - 2：法定节假日
    - 3：法定节假日前补班
    - 4：法定节假日后补班
  - 2：以 json 格式返回
- `date`：日期，格式可为`yyyy-MM-dd`、`yyyy-MM`、`yyyy`，若是后两种，只可按json格式返回。如果传入的日期解析失败，默认返回当天的信息。


json结构：

```json
{
	"code": 0,
	"message": "success",
	"data": [{
		"date": "2023-05-06",
		"name": "劳动节",
		"type": 4
	}]
}
```

其中，`code`的正常返回为`0`，异常为`-1`，同时`message`字段标明发生的错误。

## 使用的开源项目

https://github.com/NateScarlet/holiday-cn
