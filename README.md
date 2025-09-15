# miniflux-gemini

> [!WARNING]
> This is a work in progress and this documentation needs to be expanded

Expose your Miniflux instance over the Gemini protocol.

## Generate certificates

### Server side

Create and cd into `certs`, then run:

```
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout server.key -out server.crt \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=yourdomain.com"
```


### Client side

```
openssl x509 -in miniflux.crt -outform der | sha256sum
```
(or you can use the error from your gemini browser when you first attempt to connect)

```sql
INSERT INTO users(certFingerprint, instance, token)
VALUES ('ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff', 'https://mini.flux', 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbb=');
```
* `certFingerprint` is obtained at the previous step
* `instance` is the instance url, e.g. `https://minif.lux` (no need to point to any particular path)
* `token` is obtained in the Miniflux UI (Settings - API Keys)
