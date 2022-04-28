FROM nginx:alpine

RUN apk add curl

HEALTHCHECK --interval=100ms --timeout=3s CMD curl -f http://localhost/ng || exit 1
