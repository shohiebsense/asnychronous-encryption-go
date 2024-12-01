I use the CA crt both clients and server as follows

1. Generate Certificates

First, you'll need to generate certificates for both the server and the client. If you don't have a certificate authority (CA), you can create a self-signed CA to sign the server and client certificates.
Generate the CA Certificate:

# Generate CA private key
openssl genpkey -algorithm RSA -out ca.key

# Generate CA certificate
openssl req -x509 -key ca.key -out ca.crt -days 365 -subj "/CN=MyCA"

Generate the Server Certificate:

# Generate server private key
openssl genpkey -algorithm RSA -out server.key

# Create server certificate signing request (CSR)
openssl req -new -key server.key -out server.csr -subj "/CN=localhost"

# Sign the server CSR with the CA certificate
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365

Generate the Client Certificate:

# Generate client private key
openssl genpkey -algorithm RSA -out client.key

# Create client certificate signing request (CSR)
openssl req -new -key client.key -out client.csr -subj "/CN=client"

# Sign the client CSR with the CA certificate
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365
