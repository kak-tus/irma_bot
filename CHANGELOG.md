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
