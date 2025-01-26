# VSL Secrets

This is an attempt to make a user friendly, simple, accessible, suckless, no javascript, secure, zero
dependency (except the golang std library) webapp for sharing secrets.

## TLDR

Secrets are encrypted using XOR with a 1032 byte OTP, max length of secrets is 1032 bytes because of the pad length. Once viewed the secret is deleted. Secrets do not persist accross app restarts.
You may start the app with docker using:

```bash
./run_container.sh
```

or if you want to test it locally, you will need go v1.23.4, even though it should probably work with older versions as well, but I did not test this.

```bash
./run_local.sh
```

The code is minimal (~300 loc) and is supposed to be quickly inspectable by anyone with minimal golang knowledge.

## Description

The only functionality the app provides is adding a secret (max length 1032 bytes, for security reasons) and
viewing that secret a single time after which it is deleted.  
Secrets are encrypted and kept in a in memory hash map, therefore they
do not persist accross reboots.

## How it works

When you add a secret a random 1032 byte one time pad is generated, and the secret is encrypted using XOR.  
The user who added the secret is redirected and their url contains
the one time pad encoded in base64. Sharing the secret is as simple as sending someone this url. Users on that page have a "Show secret" button and after pressing it the secret will be presented to them and removed from the server.  
The limit of 1032 bytes on the secret is required because the pad is 1032 bytes, and the pad is of length 1032 bytes always, therefore by looking at the url you cannot estimate the length of the secret being shared, only that it is less than or equal to 1032 bytes.

The map key is computed as SHA512(base64OTP), therefore the user is the only one who has the one time pad decryption key.

## Why

1. I have often the need to send a sensitive piece of information to someone non tech savy, it might even be something not necessarily a secret in a cryptographic sense.
2. Not everyone knows what a pgp key is, and not everyone has a messaging app with real E2E encryption.
3. I do not want big tech spying on my messages and me getting ads for something sensitive I share with someone
4. It is much easier to send someone a link to a secret than to ask them to install a special app for that purpose.
5. OTP is theoretically the only absolutely secure encryption method, which is cool (the secret is still gonna be trasferred using https or whatever which is important to note)
6. I do not want to use a secret sharing app whose dependency tree looks like a rainforest.

## Contributing

The goal is to keep the project lean with zero dependencies and no javascript.  
I am not a security/cryptography expert and this was made with best effort, therefore it would be very much appreciated if someone could check the code for security vulnerabilities.  
Improvements to accesibility are also very much welcome as well as other general improvements.
