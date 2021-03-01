FROM node:14.4-alpine AS build

WORKDIR /usr/src/katnip
COPY . .
RUN npm install
ARG API_URI
RUN REACT_APP_API_URI=${API_URI} npm run build

EXPOSE 5000

ENTRYPOINT ["npx", "serve", "./build"]
