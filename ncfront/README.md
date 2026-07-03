# ncfront

`ncfront` is a simple reverse proxy for serving HomeDrive's Nextcloud instance,
with the required headers.

It sits in front of the Nextcloud container and forwards requests to it, while
handling a couple of Nextcloud-specific quirks that the container does not take
care of on its own:

- Redirects the service-discovery paths `/.well-known/caldav` and
  `/.well-known/carddav` to `/remote.php/dav/`, as Nextcloud expects.
- Forwards everything else to the Nextcloud backend unchanged.

## Usage

    ncfront [-nextcloud <host>] [-addr <listen-addr>]

Flags:

- `-nextcloud` — the Nextcloud host to forward to. If unset, the `NEXTCLOUD`
  environment variable is used; if that is also unset, it defaults to
  `nextcloud:80`.
- `-addr` — the address to listen on. Defaults to `:8080`.
