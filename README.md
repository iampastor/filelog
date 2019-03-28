# filelog

write file log

```go
type UserInfo struct {
	Username string
	Age      int
	Birthday time.Time
	Homepage string
	Image    []byte
}

var userInfo = &UserInfo{
	Username: "iampastor",
	Age:      12,
	Birthday: time.Now(),
	Homepage: "https://www.google.com",
	Image:    make([]byte, 1024),
}

w := filelog.NewWriter(".", "test")
data, _ := json.Marshal(userInfo)
for i := 0; i < 1; i++ {
    w.Write(data)
}
w.Close()
```