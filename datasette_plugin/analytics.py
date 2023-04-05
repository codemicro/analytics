import os
import random
import string

from datasette import hookimpl, Response
from dominate.tags import *
import yaml


MODAL_TARGET_ID = "modal-target"


with open(os.path.join(os.path.dirname(os.path.realpath(__file__)), "script.js")) as f:
    script_src = f.read()


with open("config.yml") as f:
    configs = yaml.load(f, yaml.BaseLoader)
    required_user_group = configs.get("datasette", {}).get("edit", {}).get("requiredUserGroups", "")


def is_request_authed(request):
    return is_actor_authed(request.actor)


def is_actor_authed(actor):
    if actor is None:
        return False
    return required_user_group in actor.get("groups", {})


@hookimpl
def permission_allowed(actor, action):
    if action == "execute-sql" or action == "permissions-debug" or action == "debug-menu":
        return is_actor_authed(actor)


@hookimpl
def extra_body_script(database, table, view_name, request, datasette):
    if not is_request_authed(request):
        return None

    if not (view_name == "table" and table == "config"):
        return None
    
    res = script_src.replace("{{#url}}", datasette.urls.path(datasette.urls.table(database, table) + "/-/config-editor/view"))
    res = res.replace("{{#id}}", MODAL_TARGET_ID)

    return {
        "script": res,
    }


@hookimpl
def register_routes():
    return [
        (r"^/(?P<database>.*)/(?P<table>.*)/-/config-editor/view$", Handlers.view_all),
        (r"^/(?P<database>.*)/(?P<table>.*)/-/config-editor/close$", Handlers.close_modal),
        (r"^/(?P<database>.*)/(?P<table>.*)/-/config-editor/edit/(?P<key>.*)$", Handlers.edit_record),
    ]


def generate_modal(datasette, request, config_options):
    d = div(id="editor-modal", style="width: 100%; height: 100%; position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); background-color: rgba(0, 0, 0, 0.25);")
    _d = div(style="padding: 15px; height: 280px; width: 570px; position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); background-color: white; border-radius: 0.5ch;")
    d.add(_d)

    h = h3("Edit values", close := a("(close)", href="#"))
    close["hx-get"] = datasette.urls.path(datasette.urls.table(request.url_vars["database"], request.url_vars["table"]) + "/-/config-editor/close")
    close["hx-target"] = "#" + MODAL_TARGET_ID

    _d.add(h)

    para = p()
    
    for i, opt in enumerate(config_options):
        para.add(generate_setting_span(datasette, request, opt["id"], opt["value"]))
        if i != len(config_options) - 1:
            para.add(br())

    _d.add(para)

    return d


def generate_setting_span(datasette, request, setting_id, value):
    anchor = a("(edit)", href="#")
    anchor["hx-get"] = datasette.urls.path(datasette.urls.table(request.url_vars["database"], request.url_vars["table"]) + "/-/config-editor/edit/" + setting_id)
    sp = span(code(setting_id), ": ", code(value), anchor)
    sp["hx-target"] = "this"
    return sp


def generate_edit_span(datasette, request, setting_id, current_value):
    input_id = "config-edit-" + generate_random_string(10)
    csrf_token_id = "config-edit-" + generate_random_string(10)

    entry_box = input_(type="text", name="new_value", value=current_value, _id=input_id)
    csrf_box = input_(_id=csrf_token_id, type="hidden", name="csrftoken", value=request.cookies.get("ds_csrftoken", ""))

    anchor = a("(save)", href="#")
    anchor["hx-post"] = datasette.urls.path(datasette.urls.table(request.url_vars["database"], request.url_vars["table"]) + "/-/config-editor/edit/" + setting_id)
    anchor["hx-include"] = f"#{input_id},#{csrf_token_id}"

    sp = span(code(setting_id), ": ", entry_box, anchor, csrf_box)
    sp["hx-target"] = "this"
    return sp


class Handlers:
    @staticmethod
    def not_authed_response():
        return Response(
            "Forbidden",
            status=403,
        )

    @staticmethod
    async def view_all(datasette, request):
        if not is_request_authed(request):
            return Handlers.not_authed_response()

        config_options = await datasette. \
            get_database(request.url_vars["database"]). \
            execute("""SELECT "id", "value" FROM "config";""")
        return Response.html(
            generate_modal(datasette, request, config_options).render(),
        )

    @staticmethod
    async def close_modal(request):
        if not is_request_authed(request):
            return Handlers.not_authed_response()

        return Response(
            "",
            status=204,
            headers={"hx-refresh": "true"},
        )

    @staticmethod
    async def edit_record(datasette, request):
        if not is_request_authed(request):
            return Handlers.not_authed_response()

        setting_id = request.url_vars["key"]
        current_value = await datasette. \
            get_database(request.url_vars["database"]). \
            execute("""SELECT "value" FROM "config" WHERE "id"=?;""", (setting_id,))
        current_value = list(current_value)[0]["value"]

        if request.method == "GET":
            return Response.html(
                generate_edit_span(datasette, request, setting_id, current_value).render()
            )
        elif request.method == "POST":
            vars = await request.post_vars()
            new_value = vars["new_value"]
            key = request.url_vars["key"]
            await datasette.get_database(request.url_vars["database"]).execute_write("""UPDATE "config" SET "value" = ? WHERE "id" = ?""", (new_value, key))
            return Response.html(
                generate_setting_span(datasette, request, key, new_value).render()
            )

        return Response(
            "method not allowed",
            status=405,
            content_type="text/html",
        )


def generate_random_string(n: int) -> str:
    return "".join(random.choices(string.ascii_letters, k=n))
