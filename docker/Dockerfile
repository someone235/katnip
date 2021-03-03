# -- multistage docker build: stage #1: build stage
FROM node:14.4-alpine AS build

WORKDIR /usr/src/katnip
RUN apk add yarn
COPY yarn.lock .
COPY package.json .
RUN yarn
COPY . .
ARG API_URI
RUN REACT_APP_API_URI=${API_URI} yarn run build

# --- multistage docker build: stage #2: runtime image
FROM node:14.4-alpine
WORKDIR /app

RUN apk add --no-cache tini
RUN npm i -g serve
COPY --from=build /usr/src/katnip/build /app/

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["serve"]
