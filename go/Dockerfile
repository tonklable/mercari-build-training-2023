FROM alpine

RUN addgroup -S mercari && adduser -S trainee -G mercari
# RUN chown -R trainee:mercari /path/to/db

USER trainee

# Create image directory
RUN mkdir -p /app/images

RUN CGO_ENABLED=1 GOOS=linux go build -o /mercari

CMD ["/mercari"]
