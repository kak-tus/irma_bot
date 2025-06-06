2025-05-15 v1.10.5
  - Show cutten user nickname if it is probably spammer.

2025-05-14 v1.10.4
  - Repeat some call to telegram.

2025-05-14 v1.10.3
  - Show cutten user nickname if it is probably spammer.

2025-04-15 v1.10.2
  - Temporary ban annoying user.

2024-09-19 v1.10.1
  - Bigger admin token expire timeout.

2024-07-07 v1.10.0
  - Use different urls for ui and for tg webhooks.

2024-01-14 v1.9.0
  - Use only chat system message to get newbies.
  - Remove duplicate messages.

2024-01-02 v1.8.2
  - Fix group, not exists in db, get after switch to native pgx.

2023-12-31 v1.8.1
  - Fix group get after switch to native pgx.

2023-12-30 v1.8.0
  - Fix bot stop.
  - Use versioned generators.
  - Use latest sqlc to switch to native pgx.
  - Use native pgx pool to allow more stable db connection tuning.

2023-11-11 v1.7.0
  - Remove support for set ban all users for emojii.
  - Process chat_member actions to get newbies from it.

2023-11-11 v1.6.6
  - Support for send message to all chats.

2023-11-11 v1.6.5
  - Corrrect ban count.

2023-11-11 v1.6.4
  - Read and set ban for emojii.

2023-11-11 v1.6.3
  - Support for set ban all users for emojii. This is temporary fix, because telegram does not send join chat messages.

2023-11-10 v1.6.1
  - Try fix getting join requests.

2023-11-07 v1.6.0
  - Parallel message processing.

2023-08-31 v1.5.6
  - Ban old users with first message (bigger newbie store period).

2023-08-31 v1.5.2, v1.5.3, v1.5.4, v1.5.5
  - More logs.

2023-08-21 v1.5.1,
  - More logs.
  - Use zerolog (I prefer it now) for some new log.

2023-08-21 v1.5.0
  - Fix openapi linter warnings.
  - Use latest oapi.
  - Support for ignore domains in message from newbie.
  - Go 1.21.0.
  - Ban for multiple emojii in message from newbie.

2022-01-20 v1.4.6
  - Reply to in message.

2022-01-20 v1.4.5
  - Improve message.

2022-01-20 v1.4.4
  - Send success message about config url in public chat too.

2022-01-20 v1.4.3
  - Fix broken question protection.

2022-01-20 v1.4.2
  - Fix validation.

2022-01-20 v1.4.1
  - Bring back lengths validations.

2022-01-19 v1.4.0
  - Switch to only env config.
  - Change env name from IRMA_STORAGE_REDISADDRS to IRMA_STORAGE_REDIS_ADDRS.
  - Switch to codegened database code.
  - Update all dependencies.
  - Build with ko and use github container registry.
  - Question protection enabled by default.
  - API for frontend.
  - Configuration from frontend.
  - Removed configuration by commands.

2020-10-11 v1.3.3
  - Fix start.

2020-10-11 v1.3.2
  - Update all dependencies.
  - Ban for video.

2020-10-10 v1.3.1
  - Select from db fix.

2020-10-10 v1.3.0
  - Go 1.15.
  - Support for custom ban timeout per group.

2019-11-23 v1.2.5
  - Fixed bug with error while disabling protection.
  - Fixed bug with save question settings.
  - Fixed linter warnings.

2019-10-06 v1.2.4
  - Fixed logging.

2019-10-06 v1.2.3
  - Improve logging.

2019-10-05 v1.2.2
  - Improve logging.

2019-10-05 v1.2.1
  - Remove vendoring.
  - Go 1.13.1.
  - Remove private dependency (app).

2019-09-01 v1.1.8
  - Skip checking newbies added by admins.

2019-05-19 v1.1.7
  - Revert packing. Image size note reduced significantly.

2019-05-19 v1.1.6
  - Smaller docker image size.

2019-05-02 v1.1.5
  - Add CaptionEntities support (this adds more ban options).

2019-04-26 v1.1.4
  - Correctly process already deleted messages. Not stop after "already deleted"
    error.

2019-04-16 v1.1.3
  - Remove default ban_url setting from database. It is controlled by code.
  - Lower kick pool period.
  - Fixed work with new chats with null values in db.
  - Questions protection enabled by default.
  - Special protection from immediately added messages.

2019-04-12 v1.1.2
  - Fixed kicked deletion.

2019-04-11 v1.1.1
  - Rewritten on Go from Perl to improve Memory/CPU usage and docker image size.

2019-04-06 v0.23.2
  - Fixed error logging.

2018-11-14 v0.23.1
  - Use semver versioning.
  - Fix question deletion in case of URL spam after join.

2018-11-10 0.23
  - Delete questions and join messages.

2018-11-10 0.22
  - Postgresql 10 compatibility.

2018-08-04 0.21
  - Bigger newbie detection period.

2018-08-04 0.20
  - 500 error bugfix.

2018-08-04 0.19
  - 500 error bugfix.

2018-08-04 0.18
  - Fix ban repeated logins.

2018-07-26 0.17
  - Ban users with extra long first name or last name.

2018-07-22 0.16
  - Ban by photo and link in caption.

2018-07-21 0.15
  - Delete join message on banned user.

2018-07-21 0.14
  - Ban users with extra long names.

2018-07-11 0.13
  - Healthcheck.

2018-07-11 0.12
  - Fix ban to chat names in text.

2018-05-17 0.11
  - Fix ban to links in text.

2018-04-22 0.10
  - Restore send messages to Telegram.

2017-09-06 0.9
  - Ban for restricted data in 3 first messages.

2017-08-03 0.8
  - Sticker is restricted too.

2017-07-30 0.7
  - Bugfix: users not kicked.

2017-07-27 0.6
  - Kick user with ban for 1 day.

2017-07-23 0.5
  - Fix detection messages, forwarded from chats.

2017-07-19 0.4
  - Ban by URLs support.

2017-07-16 0.3
  - Delete spammer message fix.
  - Tested in supergroup, works fine.
  - Multiprocess execution support.

2017-07-15 0.2
  - Better logging.
  - Unified prod and dev environment.

2017-07-15 0.1
  - First release.
