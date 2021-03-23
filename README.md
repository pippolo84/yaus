# yaus

yaus (**Y**et **A**nother **U**RL **S**hortener) is a simple URL shortener used to demonstrate some handy patterns to develop Go services.

It currently offers two endpoints:

1) `/shorten`: to get a "shortened" hash

```
curl --request POST \
  --url http://localhost:8080/shorten \
  --header 'Content-Type: application/json' \
  --data '{
	"url": "http://www.google.it"
}'
{"hash":"dda15434615ed3debc02fef8bbea9236"}
```

2) `/{hash}`: to get the proper redirect from the given "shortened" URL

```
$ curl --request GET \
  --url http://localhost:8080/dda15434615ed3debc02fef8bbea9236
<a href="http://www.google.it">Temporary Redirect</a>.
```