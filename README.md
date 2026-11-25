# mypass

## introduction

A terminal password manager, simple to use, less to remember, and safety enough.

## features

1. 首次使用执行 mypass init 完成初始化，密码库保存在用户主目录下的 ~/.mypass 下，设置环境变量 MYPASS_HOME 可改用其他目录，初始化时指定一个用于同步的 git 仓库地址；
2. 使用两个密钥保护数据：pincode 加密整个密码库，token 加密每一条密码；浏览列表只需要 pincode，存取密码才需要 token；两个密钥都经 Argon2id 加盐派生，数据带完整性校验，密钥输错会直接报错；
3. 每个密钥既可以是一段字符串，也可以是任意文件的内容，文件可以是本地文件或网络文件；输入密钥和密码时终端不回显，新设置的密钥和密码不能为空；
4. 每个账号以标题唯一标识，记录包含分组、标题、密码和描述；执行 mypass put 添加账号，同一标题再次添加视为修改密码，旧密码作为历史记录保留；
5. 执行 mypass get 按标题（-t）或序号（-n）取回密码，默认只复制到剪贴板、不在终端显示，指定 -p 时才打印到终端；
6. 执行 mypass history 查看某个账号的全部历史密码；
7. 可以列出所有分组、所有标题和全部账号，也可以按分组或标题模糊查询；账号列表按分组和标题排序并从 1 开始编号，每次保存时重新编号，以表格显示序号、分组、标题、创建时间和描述；
8. 执行 mypass delete 按标题删除某个账号（连同历史密码），或按分组删除整个分组，需指定 -t 或 -g，没有匹配的记录时会报错；
9. 执行 mypass old 列出超过 6 个月未修改的密码，提醒及时更换；
10. 执行 mypass key 更换密钥：验证旧 token 后，设置新的 pincode 和 token，全部密码自动重新加密；
11. 执行 mypass keygen 生成指定长度的随机密码（1 到 32 位），默认复制到剪贴板；
12. 旧格式的密码库执行 mypass upgrade 一次性升级，旧文件自动备份为 pass.v1.bak。升级沿用原来的 pincode 和 token，而 v1 的密钥派生远弱于 v2，升级后请执行 mypass key 更换密钥。

Security:

1. Use two keys to protect data, one key to encrypt each password of every records, and the other encrypt the whole store.
2. If choose key passA to protect password of record R, then the AES256 key is hmac of R.title and the key derived from passA.
3. Keys are derived from their factors by Argon2id with a random salt, data is encrypted by AES-256-GCM, so a wrong key or a changed file is told apart.
4. copy password into clipboard, your password can never be exposed.

Usability:

1. The pincode and the token can each be a string or any file, local or fetched from internet by url. People are good at remembering images not codes.
2. All history of your password will be kept.
3. Every account has a sn, `mypass get -n` finds it without its title.

## build

```
make
```

## usage

```
mypass --help
```
