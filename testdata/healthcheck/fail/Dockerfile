FROM nginx:alpine

RUN apk add curl

HEALTHCHECK --interval=5s --timeout=3s CMD curl -f http://localhost/ng || exit 1
