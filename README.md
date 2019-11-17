# Custom emoticons bot

## How to

1. Create your dictionaly to Public spreadsheet.

```csv
logo,279473996351799299/507154249705062410/unknown.png
```

2. Run the server

```bash
# Build Image
docker build yanorei32/custom-emoticons-bot ./

# Run
docker run \
	-itd \
	--init \
	--name custom-emoticons-bot \
	--restart=always \
	--memory=32mb \
	-e "APIKEY=[YOUR DISCORD API KEY]" \
	-e "DICT_URI=http://docs.google.com/spreadsheets/d/[YOUR PUBLIC SPREAD SHEET]/export?format=csv&gid=0" \
	-e "BITLY_TOKEN=[YOUR BITLY TOKEN]" \
	yanorei32/custom-emoticons-bot

# Stop
docker container stop custom-emoticons-bot
docker container rm custom-emoticons-bot
```

## Warning

This bot has some funny functions.


