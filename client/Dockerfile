FROM node
WORKDIR /app
COPY --exclude=node_modules ./client/ ./
# ARG CLIENT_PORT
# ENV VITE_PORT=$CLIENT_PORT
CMD ["bash", "-c", "npm install && npm run dev"]
# EXPOSE ${CLIENT_PORT}
