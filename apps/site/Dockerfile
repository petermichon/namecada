# Static download page. No application code: Vite hashes the stylesheet, and the
# image ships only the built `dist/`.
#
# Built and pushed to ghcr.io/petermichon/namecada-site by
# .github/workflows/site.yml; deployed from site/compose.yaml.
FROM node:24-alpine AS build
WORKDIR /site
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:1.29-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /site/dist/ /usr/share/nginx/html/
