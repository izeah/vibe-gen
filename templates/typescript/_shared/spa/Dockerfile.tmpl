FROM node:24-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --ignore-scripts
COPY . .
# Public build-time setting (the API origin); never pass secrets as VITE_* variables.
ARG VITE_API_BASE_URL=
RUN npm run build

# nginx-unprivileged runs as a non-root user on port 8080. connect-src gets the same API origin
# the bundle was built with (one setting, so the two can never disagree).
FROM nginxinc/nginx-unprivileged:alpine
ARG VITE_API_BASE_URL=
ENV API_ORIGIN=$VITE_API_BASE_URL
COPY nginx.conf /etc/nginx/templates/default.conf.template
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080
