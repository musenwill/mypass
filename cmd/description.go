package cmd

const description = `
A simple terminal password manager, simple to use, less to remember, use git as cloud synchronization, for personal usage.

Security designation:
	1. Two keys protect the data, the pincode encrypts the whole store, the token encrypts every password on its own.
	2. A key comes from its factor by Argon2id, with a random salt, and data is encrypted by AES-256-GCM. A wrong key, or any change on the store file, is told apart.
	3. A factor can be a string, or the content of a file, local or from internet.
	4. a password goes into the clipboard by default, it is printed only by "get -p" and by "history".

Usability designation:
	1. People are good at remembering images not codes, take a file as a factor.
	2. All history of your password will be kept.
	3. Every account has a sn, listed by the list commands, mypass get -n gets it without the title.
`
