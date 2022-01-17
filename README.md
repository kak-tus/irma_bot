# @irma_bot

[![Build Status](https://travis-ci.org/kak-tus/irma_bot.svg?branch=master)](https://travis-ci.org/kak-tus/irma_bot)
![CodeQL](https://github.com/kak-tus/irma_bot/workflows/CodeQL/badge.svg)

Irma AntiSpam Bot for Telegram groups

This bot was written on Perl, but for Memory/CPU usage optimisations and docker image size - rewritten on Go.

## Run migration

migrate -path ./migrations/ -database $IRMA_DB_ADDR version
