package midjourney

// Token 表名:token，存放token
type Token struct {
	Token string `gorm:"column:token;index"`
}
type Whitelist struct {
	WxId string `gorm:"column:wx_id;index"`
}
type AccessLimit struct {
	WxId  string  `gorm:"column:wx_id;index"`
	Limit float64 `gorm:"column:limit"`
}
