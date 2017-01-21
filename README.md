# ObeBot
Простой бот для Slack на го. Постит рандомную картинку из выдачи гугла, запрос берёт из прямого обращения к нему.

Для работы нужен файл prop.json с ключами вида:

`{
  "slack": "XXXXXXXXX",
  "google": "YYYYYYYYYY",
  "cse": "ZZZZZZZZZZZ"
}`

где "slack" - ключ API ботов из Slack,
"google" - ключ API Custom Search API Гугла,
"cse" - идентификатор Custom Search Engine Гугла (***обязательно нужно разрешить поиск картинок!***)
