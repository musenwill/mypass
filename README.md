# mypass

## introduction

A terminal password manager, simple to use, less to remember, and safety enough.

## features

Security:
	1. Main key is 256 bits length, and use AES to encrypt you data.
	2. Main key is made up of two factors, get composed by hmac
	3. copy password into clipboard, your password can never be exposed.

Usability:
	1. Two factors of yoru main key can be string or any file, and files can be local file or get from internet by url. Pepole are good at remembering images not codes.
	2. All history of your password will be kept.
	3. If you choose a large file as a factor, then hmac may be quite slow, make it more hard to break your another factor.

## build

```
make
```

## usage

```
mypass --help
```
