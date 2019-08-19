1. 使用 git 进行密码同步
2. 主密钥使用双因子拼接，因子可以为字符串或文件，文件可以为网络文件和本地文件
3. csv 存储密码
4. 定期提醒更新密码
5. 保存历史记录


mypass init --git [github repository address]
mypass all
> pincode
> token

mypass groups
> pincode
> token

mypass titles
> pincode
> token

mypass list
> pincode
> token

mypass filter --group [group-like] --title [title-like]
> pincode
> token

mypass delete -t [title]
> pincode
> token

mypass delete -g [group]
> pincode
> token

mypass put --group [group] --title [title] --describe [describe]
> password
> pincode
> token

mypass get [title] -p
> pincode
> token

mypass history [title]
> pincode
> token

--group, -g
--title, -t
--describe, -d
--print, -p
--help, -h
