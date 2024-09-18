# Frontend Dockerfile
FROM node:18 as build

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .

ARG REACT_APP_API_HOST_URL
ENV REACT_APP_API_HOST_URL=$REACT_APP_API_HOST_URL

# Dynamically copy all environment variables ending with _JS_SCRIPT
# Use a shell script to dynamically set environment variables
RUN echo '#!/bin/sh' > /tmp/env.sh && \
    env | grep _JS_SCRIPT | while read -r line; do \
    echo "export $line" >> /tmp/env.sh; \
    done && \
    chmod +x /tmp/env.sh

# Source the environment variables and run the build
RUN . /tmp/env.sh && npm run build

FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]