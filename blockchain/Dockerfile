FROM node:20-alpine

WORKDIR /app

COPY package.json ./
RUN npm install

COPY hardhat.config.js ./
COPY contracts ./contracts
COPY scripts ./scripts
COPY wait-for-node.js docker-entrypoint.sh ./
RUN chmod +x docker-entrypoint.sh

EXPOSE 8545

# Healthy only once the node accepts RPC calls AND the contract deploy step has
# run — adapter startup depends on this so it doesn't race the deploy.
HEALTHCHECK --interval=5s --timeout=3s --start-period=20s --retries=20 \
  CMD test -f /tmp/deployed && node wait-for-node.js --once

CMD ["./docker-entrypoint.sh"]
