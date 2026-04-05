FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
ARG VITE_API_URL
ARG VITE_USE_WORKOUT_MOCKS=false
ENV VITE_API_URL=${VITE_API_URL}
ENV VITE_USE_WORKOUT_MOCKS=${VITE_USE_WORKOUT_MOCKS}
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 3000
CMD ["nginx", "-g", "daemon off;"]