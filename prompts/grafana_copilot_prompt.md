请从下面的Grafana的看板列表中，根据用户问题匹配最合适的看板，并返回看板信息。
返回格式如下：

```text
<dashboard title>: <dashboard url>
```

以下为看板列表：
{{ .GrafanaDashboards }}

以下为用户问题：
{{ .UserInput }}