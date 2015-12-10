# You may add here your
# server {
#       ...
# }
# statements for each of your virtual hosts to this file

##
# You should look at the following URL's in order to grasp a solid understanding
# of Nginx configuration files in order to fully unleash the power of Nginx.
# http://wiki.nginx.org/Pitfalls
# http://wiki.nginx.org/QuickStart
# http://wiki.nginx.org/Configuration
#
# Generally, you will want to move this file somewhere, and start with a clean
# file but keep this around for reference. Or just disable in sites-enabled.
#
# Please see /usr/share/doc/nginx-doc/examples/ for more detailed examples.
##

# another virtual host using mix of IP-, name-, and port-based configuration
#
#server {
#       listen 8000;
#       listen somename:8080;
#       server_name somename alias another.alias;
#       root html;
#       index index.html index.htm;
#
#       location / {
#               try_files $uri $uri/ =404;
#       }
#}

upstream api_server {
  server 127.0.0.1:8080; #default port
  keepalive 128;
}

# HTTPS server
#
#server {
#       listen 443;
#       server_name localhost;
#
#       root html;
#       index index.html index.htm;
#
#       ssl on;
#       ssl_certificate cert.pem;
#       ssl_certificate_key cert.key;
#
#       ssl_session_timeout 5m;
#
#       ssl_protocols SSLv3 TLSv1 TLSv1.1 TLSv1.2;
#       ssl_ciphers "HIGH:!aNULL:!MD5 or HIGH:!aNULL:!MD5:!3DES";
#       ssl_prefer_server_ciphers on;
#
#       location / {
#               try_files $uri $uri/ =404;
#       }
#}

  server {
    listen 80;
    server_name enu.io;
    return 301 https://enu.io/$request_uri;
  }

  server {
    listen 443 ssl;
    server_name enu.io;

    # SSL config goes here (removed for brevity)
ssl_certificate      /etc/nginx/ssl/nginx.crt;
  ssl_certificate_key  /etc/nginx/ssl/nginx.key;
 # support FS, and BEAST protection - https://coderwall.com/p/ebl2qa
  # SSLv3 not supported, due to poodlebleed bug
  server_tokens off;
  ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
  ssl_prefer_server_ciphers on;
  ssl_session_timeout 5m;
# ssl_ciphers 'AES128+EECDH:AES128+EDH';
ssl_ciphers "ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA:ECDHE-RSA-AES128-SHA:DHE-RSA-AES256-SHA256:DHE-RSA-AES128-SHA256:DHE-RSA-AES256-SHA:DHE-RSA-AES128-SHA:ECDHE-RSA-DES-CBC3-SHA:EDH-RSA-DES-CBC3-SHA:AES256-GCM-SHA384:AES128-GCM-SHA256:AES256-SHA256:AES128-SHA256:AES256-SHA:AES128-SHA:DES-CBC3-SHA:HIGH:!aNULL:!eNULL:!EXPORT:!DES:!MD5:!PSK:!RC4";
    add_header Strict-Transport-Security max-age=31536000;

    location / {
      proxy_pass http://api_server;
      proxy_redirect off;

      # Security headers removed, but think about X-Frame-Options, Content-Security-Policy, etc

      # Enable HTTP keep-alives
      proxy_http_version 1.1;
      proxy_set_header Connection "";

      # Buffers
      # Buffers should be greater than the mean response size to allow effective caching
      proxy_buffering on;
      proxy_buffers 32 16k;
      proxy_buffer_size 32k;
      proxy_busy_buffers_size 64k;
      proxy_temp_file_write_size 64k;

      # Pass scheme and remote host IP to proxied application
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Scheme $scheme;
      proxy_set_header Referer $http_referer;
      proxy_set_header Host $http_host;
    }
  }
