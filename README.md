## Reddit Sample Application

This application uses provided credentials to interact with Reddit's free API. It intends to 
honor rate limiting restrictions and simply prints a summary of what it collects to stdout.

### Configuration

This application has several required and optional parameters that can be injected when 
the program is started.

To get started, duplicate `.env_sample` as `.env` and set the clientID, secret, username and 
password parameters. If you prefer, you can override the parameters in the `.env` using 
environment variables.

```shell
# Run using an .env you created based off of .env_sample
make run
  
# Run using environment variables to override .env
REDDIT_CLIENT_ID=123 \
  go run cmd/reddit/main.go
```

### Makefile

See `make help`.

### Requirements

See [requirements](./requirements.md).

