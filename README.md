<div align="center">
<img src="webs/src/assets/logo.png" width="150px" height="150px" />
</div>

<div align="center">
[![Vue Version](https://img.shields.io/github/go-mod/Vue-version/xeefei/sublink.svg?style=for-the-badge)](#)
[![](https://img.shields.io/github/v/release/xeefei/sublink.svg?style=for-the-badge)](https://github.com/xeefei/sublink/releases)
[![Element Plus Version](https://img.shields.io/github/go-mod/Element-version/xeefei/sublink.svg?style=for-the-badge)](#)
[![GO Version](https://img.shields.io/github/go-mod/go-version/xeefei/sublink.svg?style=for-the-badge)](#)
[![Downloads](https://img.shields.io/github/downloads/xeefei/sublink/total.svg?style=for-the-badge)](https://github.com/xeefei/sublink/releases/latest)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?longCache=true&style=for-the-badge)]
    <a href="https://t.me/XUI_CN" target="_blank">
        <img src="https://img.shields.io/badge/TG-交流群-orange.svg"/>
    </a>
    <div align="center"> 中文 | <a href="README.en-US.md">English</div>
</div>

## [项目简介]

1、项目基于sublink项目二次开发：https://github.com/gooaclok819/sublinkX

2、前端基于：https://github.com/youlaitech/vue3-element-admin

3、后端采用go+gin+gorm，默认账号admin 密码123456，请进入后台自行修改

4、因为重写目前还有很多布局结构以及功能稍少，也需要额外花费不少时间。

## [项目特色]

1、自由度和安全性较高，能够记录访问订阅，配置轻松，

2、二进制编译直接脚本进行安装，无需Docker容器，

3、目前仅支持客户端：v2ray clash surge，

4、对于 v2rayN/v2rayNG 则为base64通用格式，

5、clash支持协议:ss ssr trojan vmess vless hy hy2 tuic，

6、surge支持协议:ss trojan vmess hy2 tuic。

## [项目预览]

![1712594176714](webs/src/assets/1.png)
![1712594176714](webs/src/assets/2.png)

## [v2.1 更新说明]

#### 后端更新

1. 修复底层代码
2. 修复各种奇葩bug
3. 建议卸载数据库(记得备份数据) 新数据库结构有些不一样可能会导致一些bug

#### 前端更新

1. 完善node页面




## [安装说明]
### linux方式：
```
curl -s -H "Cache-Control: no-cache" -H "Pragma: no-cache" https://raw.githubusercontent.com/xeefei/sublink/main/install.sh | sudo bash
```

```sublink``` 呼出菜单

然后输入安装脚本即可

### docker方式：

在自己需要的位置创建一个目录比如mkdir sublinkx

然后cd进入这个目录，输入下面指令之后数据就挂载过来

需要备份的就是db和template
```
docker run --name sublink -p 8000:8000 \
-v $PWD/db:/app/db \
-v $PWD/template:/app/template \
-v $PWD/logs:/app/logs \
-d xeefei/sublink
```


## Stargazers over time
[![Stargazers over time](https://starchart.cc/xeefei/sublink.svg?variant=adaptive)](https://starchart.cc/xeefei/sublink)

