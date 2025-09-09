MailHog API v1
==============

The v1 API is a RESTful HTTP JSON API.

### GET /api/v1/events

Streams new messages using EventSource and chunked encoding

### GET /api/v1/messages

Lists all messages excluding message content

### DELETE /api/v1/messages

Deletes all messages

Returns a ```200``` response code if message deletion was successful.

### GET /api/v1/messages/{ message_id }

Returns an individual message including message content

### DELETE /api/v1/messages/{ message_id }

Delete an individual message

Returns a ```200``` response code if message deletion was successful.

### GET /api/v1/messages/{ message_id }/download

Download the complete message

### GET /api/v1/messages/{ message_id }/mime/part/{ part_index }/download

Download a MIME part

### POST /api/v1/messages/{ message_id }/release

Release the message to an SMTP server

Send a JSON body specifying the recipient, SMTP hostname and port number:

```json
{
	"Host": "mail.example.com",
	"Post": "25",
	"Email": "someone@example.com"
}
```

Returns a ```200``` response code if message delivery was successful.

## Health and Monitoring Endpoints

The following endpoints are available for health checks and monitoring:

### GET /health

Liveness probe endpoint for Kubernetes health checks.

Returns application status, uptime, and version information.

### GET /ready

Readiness probe endpoint for Kubernetes health checks.  

Returns service readiness status and storage connectivity information.

Returns a ```200``` response code if the service is ready, or ```503``` if not ready.

### GET /metrics

Prometheus-compatible metrics endpoint.

Returns application metrics in Prometheus exposition format including:
- Message count
- Application uptime  
- Memory usage
- Active goroutines
- Storage information

See [Health Endpoints](HEALTH.md) for detailed documentation.
