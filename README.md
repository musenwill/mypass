# mypass

## introduction

A terminal password manager, simple to use, less to remember, and safety enough.

## features

Security:

1. Use two keys to protect data, one key to encrypt each password of every records, and the other encrypt the whole store.
2. If choose key passA to protect password of record R, then the AES256 key is hmac of R.title and passA.
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
