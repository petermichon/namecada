# Static download page. No build step: the files are the artifact.
#
# Built and pushed to ghcr.io/petermichon/namecada-site by
# .github/workflows/site.yml; deployed from site/compose.yaml.
FROM nginx:1.29-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY index.html style.css favicon.svg /usr/share/nginx/html/
