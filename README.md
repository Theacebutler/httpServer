# httpServer

An Http server built in go

> WARNING! This is a project made for learning purposes,
> it lacks many security features, and it is not to be used in production or
> and real word applications. If you choose to use this project,
> please be aware of the risks, or just use Go's builtin `http` library.

## Description

The server is made of two parts:

<!-- markdownlint-disable MD013 -->

1. A parser, it follows the specifications specified in [RFC 9112](https://datatracker.ietf.org/doc/html/rfc9112) and in [RFC 9110](https://datatracker.ietf.org/doc/html/rfc9110).
2. A simple http/1.1 server built on tcp using the builtin `tcp` library
   from the go standard library.
