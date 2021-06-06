# 💉 berlin-vaccine-alert

Telegram bot to get information on vaccine availabilities. Get the latest appointments directly on telegram.

The bot listen to update from doctolib doctors and clinics and clinics that have their own online booking system.

Get it there [@covid_eleboucher_bot](https://t.me/covid_eleboucher_bot)

## Usage

Create a [telegram bot](https://core.telegram.org/bots#creating-a-new-bot) to get the telegram token.

Rename `.config.example.yml` to `.config.yml` and add your token in this file.


### Local

This project use golang and sqlite3 make sure it is installed before following the next steps (unless you use docker).

Compile the project using:

```golang
go build
```

run the database migration by installing:

```bash
go get -v github.com/rubenv/sql-migrate/...
```

and run the migration:

```
sql-migrate up -env production
```

and then start the project with:

```
./covid run
```

To start receiving notification start a discussion with your newly created bot.

### Docker

if you want to use Docker instead. run the following command.

```bash
docker compose up -d
```

## Feedback

If you have any idea or feedback, feel free to create a ticket or reply to this post https://www.reddit.com/r/berlinvaccination/comments/np81h5/telegram_bot_to_get_a_vaccine_appointment/

## Support

This project is still open source feel free to open merge request. Otherwise you can support me:

- Using PayPal: https://paypal.me/ELeboucher
- Buy me a beer 🍺: https://www.buymeacoffee.com/eleboucher

## Contact

You can contact me on twitter [@elebouch](https://twitter.com/elebouch) or on telegram @genesixx
