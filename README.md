# glog

高并发日志记录组件
日志服务器不可用时，打印到文件中，按天分
可以开启debug模式，可以选择打印栈

# 安装:
- go get github.com/nbvghost/glog

# 用法：
- glog.NewLogger(logserver,"sdf",true,false)
-- logserver:日志服务器地址，glog把日志post到这个地址上，由开发者自己保存

# post例子：
``` http
  POST /log/server HTTP/1.1
  Host: www.abc.com
  Content-Type: application/json
  Cache-Control: no-cache
  
  {"JSON":"text"}
```
