# This file is derived from Simon Willison's datasette-auth0 project, which can
# be found at https://github.com/simonw/datasette-auth0 and is licensed under
# the Apache License 2.0.
#
# It has been modified to be more general - to read from the OpenID
# Connect-standard `/.well-known/openid-configuration` URL to fetch
# configuration information. As part of this, some route names, configuration
# parameters and other strings have also been changed.

from datasette import hookimpl, Response
from urllib.parse import urlencode
import baseconv
import httpx
import secrets
import time
from threading import Lock


_openid_config = None
_openid_config_lock = Lock()


def _get_openid_config(issuer):
    global _openid_config

    _openid_config_lock.acquire()

    if _openid_config is not None:
        _openid_config_lock.release()
        return _openid_config

    config_url = issuer + "/.well-known/openid-configuration"
    response = httpx.get(config_url, follow_redirects=True)
    if response.status_code != 200:
        raise ConfigError("Could not fetch OpenID configuration information: {}".format(response.status_code))
    response_json = response.json()

    _openid_config = response_json
    _openid_config_lock.release()

    return response_json


async def oidc_login(request, datasette):
    redirect_uri = datasette.absolute_url(
        request, datasette.urls.path("/-/oidc-callback")
    )

    try:
        config = _config(datasette)
    except ConfigError as e:
        return _error(datasette, request, str(e))

    try:
        openid_config = _get_openid_config(config.get("issuer"))
    except ConfigError as e:
        return _error(datasette, request, str(e))

    state = secrets.token_hex(16)
    url = openid_config.get("authorization_endpoint") + "?" + urlencode(
        {
            "response_type": "code",
            "client_id": config["client_id"],
            "redirect_uri": redirect_uri,
            "scope": config.get("scope") or "openid profile email",
            "state": state,
        }
    )
    response = Response.redirect(url)
    response.set_cookie("oidc-state", state, max_age=3600)
    return response


async def oidc_callback(request, datasette):
    try:
        config = _config(datasette)
    except ConfigError as e:
        return _error(datasette, request, str(e))
    code = request.args["code"]
    state = request.args.get("state") or ""
    # Compare state to their cookie
    expected_state = request.cookies.get("oidc-state") or ""
    if not state or not secrets.compare_digest(state, expected_state):
        return _error(
            datasette,
            request,
            "state check failed, your authentication request is no longer valid",
        )

    try:
        openid_config = _get_openid_config(config.get("issuer"))
    except ConfigError as e:
        return _error(datasette, request, str(e))

    # Exchange the code for an access token
    response = httpx.post(
        openid_config.get("token_endpoint"),
        data={
            "grant_type": "authorization_code",
            "redirect_uri": datasette.absolute_url(
                request, datasette.urls.path("/-/oidc-callback")
            ),
            "code": code,
        },
        auth=(config["client_id"], config["client_secret"]),
        follow_redirects=True,
    )
    if response.status_code != 200:
        return _error(
            datasette,
            request,
            "Could not obtain access token: {}".format(response.status_code),
        )
    # This should have returned an access token
    access_token = response.json()["access_token"]
    # Exchange that for the user info
    profile_response = httpx.get(
        openid_config.get("userinfo_endpoint"),
        headers={"Authorization": "Bearer {}".format(access_token)},
        follow_redirects=True,
    )
    if profile_response.status_code != 200:
        return _error(
            datasette,
            request,
            "Could not fetch profile: {}".format(response.status_code),
        )

    profile_json = profile_response.json()

    if "id" not in profile_json:
        profile_json["id"] = profile_json.get("sub")

    # Set actor cookie and redirect to homepage
    redirect_response = Response.redirect("/")
    expires_at = int(time.time()) + (24 * 60 * 60)
    redirect_response.set_cookie(
        "ds_actor",
        datasette.sign(
            {
                "a": profile_json,
                "e": baseconv.base62.encode(expires_at),
            },
            "actor",
        ),
    )
    return redirect_response


@hookimpl
def register_routes():
    return [
        (r"^/-/oidc-login$", oidc_login),
        (r"^/-/oidc-callback$", oidc_callback),
    ]


class ConfigError(Exception):
    pass


def _config(datasette):
    config = datasette.plugin_config("oidc.py")
    missing = [
        key for key in ("issuer", "client_id", "client_secret") if not config.get(key)
    ]
    if missing:
        raise ConfigError(
            "The following oidc plugin settings are missing: {}".format(
                ", ".join(missing)
            )
        )
    return config


def _error(datasette, request, message):
    datasette.add_message(request, message, datasette.ERROR)
    return Response.redirect("/")


@hookimpl
def menu_links(datasette, actor):
    if not actor:
        return [
            {
                "href": datasette.urls.path("/-/oidc-login"),
                "label": "Sign in with OpenID Connect",
            },
        ]
