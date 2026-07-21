# Using cargo-chef for faster builds in Docker
# See: https://github.com/LukeMathWalker/cargo-chef
FROM lukemathwalker/cargo-chef:latest-rust-1.77 AS chef
WORKDIR /app

FROM chef AS planner
COPY . .
RUN cargo chef prepare --recipe-path recipe.json

FROM chef AS builder 
COPY --from=planner /app/recipe.json recipe.json
# Build dependencies - this is the caching layer!
# This layer will be cached as long as Cargo.toml/Cargo.lock don't change.
RUN cargo chef cook --release --recipe-path recipe.json
# Now build the actual application
COPY . .
RUN cargo build --release --bin broker

# Final runtime image
FROM debian:bookworm-slim AS runtime
WORKDIR /app
# Copy the compiled binary from the builder stage
COPY --from=builder /app/target/release/broker /app/broker

# The broker service is reachable on port 3400
EXPOSE 3400

# Run the broker
ENTRYPOINT ["./broker"]
