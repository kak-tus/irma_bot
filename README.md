# @irma_bot

[![Test](https://github.com/kak-tus/irma_bot/actions/workflows/test.yml/badge.svg)](https://github.com/kak-tus/irma_bot/actions/workflows/test.yml)
[![Build image](https://github.com/kak-tus/irma_bot/actions/workflows/build-image.yml/badge.svg)](https://github.com/kak-tus/irma_bot/actions/workflows/build-image.yml)
[![CodeQL](https://github.com/kak-tus/irma_bot/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/kak-tus/irma_bot/actions/workflows/codeql-analysis.yml)

Irma AntiSpam Bot for Telegram groups

This bot was written on Perl, but for Memory/CPU usage optimisations and docker image size - rewritten on Go.

## Run migration

migrate -path ./migrations/ -database $IRMA_DB_ADDR version
