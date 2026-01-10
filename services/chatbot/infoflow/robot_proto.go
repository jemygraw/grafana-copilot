package infoflow
 
const (
	MessageBodyTypeText     = "TEXT"
	MessageBodyTypeLink     = "LINK"
	MessageBodyTypeAt       = "AT"
	MessageBodyTypeImage    = "IMAGE"
	MessageBodyTypeMarkdown = "MD"
)

type RequestBody struct {
	Message Message `json:"message"`
}

type MessageOptions struct {
	AtUserIds []string
	AtAll     bool
}

func (m *MessageOptions) IsAtEnabled() bool {
	return m.AtAll || len(m.AtUserIds) > 0
}

func (m *MessageOptions) CreateAtBody() (atBody MessageBody) {
	atBody = MessageBody{
		Type: MessageBodyTypeAt,
	}
	if m.AtAll {
		atBody.AtAll = true
	} else {
		atBody.AtUserIds = m.AtUserIds
	}
	return
}

type ResponseBody struct {
	ErrorCode    int       `json:"errcode"`
	ErrorMessage string    `json:"errmsg"`
	Data         ExtraData `json:"data"`
}

type ExtraData struct {
	Fail map[string]int `json:"fail"`
}

type Message struct {
	Header MessageHeader `json:"header"`
	Body   []MessageBody `json:"body"`
}

type MessageHeader struct {
	ToId []int `json:"toid"`
}

type MessageBody struct {
	Type      string   `json:"type"`
	Content   string   `json:"content,omitempty"`
	Href      string   `json:"href,omitempty"`
	AtUserIds []string `json:"atuserids,omitempty"`
	AtAll     bool     `json:"atall,omitempty"`
}

const ErrNone = 0

var ErrorMap = map[int]string{
	-1:    "系统错误",
	40000: "参数错误",
	40035: "群聊ID不合法",
	40036: "群聊非企业群",
	40040: "参数错误",
	40044: "机器人未被添加到群中",
	40045: "agentId不合法",
	40046: "发送消息频率超限",
	40047: "文件上传失败",
	40060: "body超过9k",
	40061: "text类型文本总长度超过2k",
	40062: "link类型单个链接长度超过1k",
	40063: "image类型图片数量超过1个",
	40064: "header中offlinenotify长度超过1k",
	40065: "header中compatible长度超过1k",
	40066: "image类型图片大小超过1m",
	40067: "markdown类型数量超过1个",
	40068: "markdown内容长度超过2048个字符",
	40069: "message属性格式不正确",
	40071: "at超过50人",
	40200: "机器人发送消息权限已被封禁",
	40201: "机器人接受消息权限已被封禁",
	40300: "机器人已被停用",
}

func GetErrorMessage(code int) (err string) {
	return ErrorMap[code]
}
