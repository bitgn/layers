def page(nav, body):
    return """<!DOCTYPE html>
<head>
<meta charset="utf-8">
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-beta.2/css/bootstrap.min.css" integrity="sha384-PsH8R72JQ3SOdhVi3uxftmaW6Vc51MKb0q5P2rRUpPvrszuE4W1povHYgTpBfshb" crossorigin="anonymous">

</head>
<body>"""+nav+"""



<div class="container">"""+body+'</div></body></html>';



def img(src):
    return '<img src="' + src + '" class="img-fluid"/>';


def sect(inner):
   return '<div class="row"><div class="col m-3">\n' + inner + '\n</div></div>\n';

def h1(text):
    return '<h1>' + text + '</h1>';


def simple_sect(title,*fields):
    body = "<h1>" + title + "</h1>"

    for f in fields:
        if f.endswith(".png"):
            body += img(f)
            continue
        body += "<p>" +f +"</p>"

    return sect(body)

def nav_bar(title, *links):


    body = """<nav class="navbar navbar-expand-lg navbar-dark bg-dark">
  <a class="navbar-brand" href="#">"""+title+"""</a>

<div class="collapse navbar-collapse" id="navbarSupportedContent">
    <ul class="navbar-nav ml-auto">"""

    for ref, title in links:
        body += '<li class="nav-item active">'
        body += '<a class="nav-link" href="' + ref + '">' + title +'</a>'
        body += '</li>'
    body += '</ul></div></nav>'
    return body
