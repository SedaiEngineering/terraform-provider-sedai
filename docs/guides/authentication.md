# Sedai authentication

## Authentication using environment variables

The Sedai provider requires the following environment variables:
- `SEDAI_API_TOKEN`
- `SEDAI_BASE_URL`

## Authentication using .env file

The Sedai provider also supports reading from a `.env` file. The file have the following content:

```
  SEDAI_API_TOKEN=sedai_api_token
  SEDAI_BASE_URL=https://env.sedai.app
```
