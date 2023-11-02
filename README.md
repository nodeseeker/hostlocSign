# hostlocSign
 一款基于Golang的hostloc签到工具，支持多用户和telegram通知。



## 使用教程



### 查找对应文件

打开程序的发布页 [https://github.com/nodeseeker/hostlocSign/releases](https://github.com/nodeseeker/hostlocSign/releases)，在列表中找到对应CPU架构的程序（如下图），比如x86_64的Linux系统，即为`hostlocSign-linux-amd64.zip`。



### 创建文件路径

在Linux系统中，使用`root`用户权限，执行以下命令创建文件路径：

```shell
mkdir /opt/hostlocSign
cd /opt/hostlocSign
```

下载文件并解压，其中包含二进制文件`hostlocSign`和`config`配置文件：

```shell
wget https://github.com/nodeseeker/hostlocSign/releases/download/v1.0.0/hostlocSign-linux-amd64.zip
unzip hostlocSign-linux-amd64.zip
```



### 修改配置文件

配置文件`config.json`的内容如下，添加自己的用户信息。
```json
{
  "sleep_time": 30,
  "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
  "telegram": {
    "enable": false,
    "token": "123456789:ABCDEFG-ZXCVBNM",
    "chat_id": "9876543210"
  },
  "accounts": [
    {"username": "cpuer", "password": "hostloc123456789"},
    {"username": "admin", "password": "123456789hostloc"}
  ]
}
```

其中：

1. `sleep_time`是两次访问页面的间隔时间，不建议改小，否则可能会引发cc验证。
2. `user_agent`模仿浏览器，如果熟悉UA可以更改为自己常用的，负责不建议更改。
3. `telegram`中的`enable`为是否开启电报推送（默认关闭），改成`true`则会在签到出错的时候给telegram发通知；`token`和`chat_id`顾名思义。
4. `accounts`中为用户名和密码，支持多账户。如果只有一个账户，则删除**第一行**；如果有更多账户，则在现有的两个账户中按照第一行的格式添加。注意最后一个用户的信息结尾没有`,`号，遵循`json`格式要求。



### 定时运行签到

使用`crontab`进行定时签到，终端中输入以下内容：

```shell
crontab -e
```

输入以下内容，实现定时运行程序进行签到：

```shell
0 2 * * * /opt/hostlocSign/hostlocSign >> /opt/hostlocSign/error.log 2>&1
```

上述为每天凌晨2点运行一次签到，将获得20个积分（如果是当日首次登陆，将会额外获取1积分）。

程序将会在`/opt/hostlocSign`文件夹下生成两个新的文件：

1. `scores.log` 记录执行的时间和具体积分
2. `error.log` 为报错记录

