# go-favzap

**go-favzap** is an application to chat via WhatsApp with only one person, your favorite person.

This project uses *go-whatsapp* developed by [Rhymen](https://github.com/Rhymen/go-whatsapp)

## Installation
```sh
go get github.com/Rhymen/go-whatsapp
```
```sh
go get github.com/dhinojosac/go-favzap
```
### Windows
```sh
make build-win
```
### Linux
```sh
make build-linux
```

## Usage 

### Windows
```sh
./go-favzap.exe <whatsapp-contact-number>
```
### Linux
```sh
./go-favzap <whatsapp-contact-number>
```
It will first load the history messages. (bug: sometimes does not show them in order).

To add audio notifications send command ```*s```.

