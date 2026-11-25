1. 使用 git 进行密码同步
2. 使用 pincode 和 token 两个密钥，每个密钥的因子可以为字符串或文件，文件可以为网络文件和本地文件
3. csv 存储密码
4. 定期提醒更新密码
5. 保存历史记录


mypass init --git [github repository address]

mypass all
> pincode

mypass groups
> pincode

mypass titles
> pincode

mypass filter --group [group-like] --title [title-like]
> pincode

mypass delete -t [title]
> pincode

mypass delete -g [group]
> pincode

mypass put --group [group] --title [title] --describe [describe]
> pincode
> token
> password

mypass get -t [title] -p
> pincode
> token

mypass get -n [sn]
> pincode
> token

mypass history -t [title]
> pincode
> token

mypass key
> pincode
> token
> new pincode
> new pincode again
> new token
> new token again

--group, -g
--title, -t
--sn, -n
--describe, -d
--print, -p
--help, -h
