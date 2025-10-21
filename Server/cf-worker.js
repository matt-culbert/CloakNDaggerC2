addEventListener('fetch', event => {
  event.respondWith(handleRequest(event.request));
});

const STORED_TRANSACTION_ID = new Uint8Array([0x04, 0xd2]);

async function handleRequest(request) {
  // Get the raw DNS query (Base64-encoded DNS message) from the request body
  const rawBody = await request.arrayBuffer();
  const dnsQueryBuffer = new Uint8Array(rawBody);

  try {
    // Extract the transaction ID (first two bytes) and domain name
    const transaction_id = extractTransactionId(dnsQueryBuffer);
    const domain = extractDomainNameFromQuery(dnsQueryBuffer);

    const isMatch = compareTransactionId(transaction_id, STORED_TRANSACTION_ID);
    if (!isMatch) {
      return createNxDnsResponse(domain, transaction_id);
    }
    const response = await fetch('http://127.0.0.1:5000/1234');
    const ipv6List = await response.json();

    return createAnswerDnsResponse(domain, transaction_id, ipv6List);

  } catch (error) {
    return new Response({ status: 500 });
  }
}

function createNxDnsResponse(domain, transactionId) {
  // Modify the Opcode to be something other than 0 (standard query)
  // Opcode is 4 bits in the Flags field (bits 1-4)
  const qr = 1;           // This is a response (QR bit is 1)
  const opcode = 0b0010;   // Opcode for 'Server status request' (2)
  const aa = 0;            // Not authoritative (AA bit)
  const tc = 0;            // Not truncated (TC bit)
  const rd = 1;            // Recursion desired (RD bit)
  
  // Construct the first byte of the flags
  const flags1 = (qr << 7) | (opcode << 3) | (aa << 2) | (tc << 1) | rd;

  const ra = 1;  // Recursion available
  const z = 0;   // Reserved (Z bit)
  const ad = 0;  // Authentic data
  const cd = 0;  // Checking disabled
  const rcode = 3;  // RCODE for "NXDOMAIN" (3)

  // Construct the second byte of the flags
  const flags2 = (ra << 7) | (z << 6) | (ad << 5) | (cd << 4) | rcode;

  // Flags (2 bytes)
  const flags = new Uint8Array([flags1, flags2]);

  const qdcount = new Uint8Array([0x00, 0x01]); // 1 question
  const ancount = new Uint8Array([0x00, 0x00]); // No answers
  const nscount = new Uint8Array([0x00, 0x00]); // No authority records
  const arcount = new Uint8Array([0x00, 0x00]); // No additional records
  
  // Encode the domain name in DNS label format
  const encodedDomain = encodeDomainName(domain);

  // Build the question section
  const question = new Uint8Array([
    ...encodedDomain,      // Encoded domain name
    0x00, 0x01,            // Type A (0x0001)
    0x00, 0x01             // Class IN (0x0001)
  ]);

  // Concatenate everything into a DNS response
  const response = new Uint8Array([
    ...transactionId, ...flags, ...qdcount, ...ancount, ...nscount, ...arcount, ...question
  ]);

  return new Response(response.buffer, {
    headers: { 'Content-Type': 'application/dns-message' }
  });
}

function createAnswerDnsResponse(domain, transaction_id, ipv6List) {
  const qr = 1;            // This is a response (QR bit is 1)
  const opcode = 0b0000;   // Standard opcode
  const aa = 1;            // Authoritative (AA bit)
  const tc = 0;            // Not truncated (TC bit)
  const rd = 1;            // Recursion desired (RD bit)

  const flags1 = (qr << 7) | (opcode << 3) | (aa << 2) | (tc << 1) | rd;
  const ra = 1;  // Recursion available
  const z = 0;   // Reserved (Z bit)
  const ad = 1;  // Authentic data
  const cd = 0;  // Checking disabled
  const rcode = 0;
  const flags2 = (ra << 7) | (z << 6) | (ad << 5) | (cd << 4) | rcode;

  const flags = new Uint8Array([flags1, flags2]);

  const qdcount = new Uint8Array([0x00, 0x01]); // 1 question
  const ancount = new Uint8Array([0x00, ipv6List.length]); // Number of answers
  const nscount = new Uint8Array([0x00, 0x00]); // No authority records
  const arcount = new Uint8Array([0x00, 0x00]); // No additional records

  const encodedDomain = encodeDomainName(domain);

  const question = new Uint8Array([
    ...encodedDomain,      // Encoded domain name
    0x00, 0x1C,            // Type AAAA (IPv6)
    0x00, 0x01             // Class IN (0x0001)
  ]);

  const answers = ipv6List.map(ipv6 => {
    const ipv6Bytes = parseIPv6(ipv6);
    const rdLength = new Uint8Array([0x00, 0x10]); // Length of IPv6 address

    return new Uint8Array([
      ...encodedDomain,   // Domain name in label format
      0x00, 0x1C,         // Type AAAA
      0x00, 0x01,         // Class IN
      0x00, 0x00, 0x00, 0x3C,  // TTL (60 seconds)
      ...rdLength,        // Length of data (16 bytes for IPv6)
      ...ipv6Bytes        // IPv6 address bytes
    ]);
  });

  const flattenedAnswers = answers.reduce(
    (acc, answer) => acc.concat(Array.from(answer)),
    []
  );

  const response = new Uint8Array(
    transaction_id.length +
    flags.length +
    qdcount.length +
    ancount.length +
    nscount.length +
    arcount.length +
    question.length +
    flattenedAnswers.length
  );

  let offset = 0;
  response.set(transaction_id, offset); offset += transaction_id.length;
  response.set(flags, offset); offset += flags.length;
  response.set(qdcount, offset); offset += qdcount.length;
  response.set(ancount, offset); offset += ancount.length;
  response.set(nscount, offset); offset += nscount.length;
  response.set(arcount, offset); offset += arcount.length;
  response.set(question, offset); offset += question.length;
  response.set(flattenedAnswers, offset);

  console.log("Response Length:", response.length);
  return new Response(response.buffer, {
    headers: { 'Content-Type': 'application/dns-message' }
  });
}


// Helper function to expand and convert IPv6 address to 16-byte array
function expandAndConvertIPv6(ipv6) {
  const expandedSegments = ipv6.split('::').map(part => part.split(':'));
  let segments = expandedSegments[0];

  // Add empty segments if '::' was found to fill 8 parts total
  if (expandedSegments.length === 2) {
    const suffix = expandedSegments[1];
    const missingSegments = 8 - (segments.length + suffix.length);
    segments = segments.concat(Array(missingSegments).fill('0'), suffix);
  }

  return segments.flatMap(part => {
    const value = parseInt(part || '0', 16); // Ensure zero for empty parts
    return [(value >> 8) & 0xff, value & 0xff]; // Split 16-bit segment into two bytes
  });
}

// Helper function to make the AAAA answers section of the DNS response
function constructAAAAAnswer(domain) {
  const labels = domain.split('.');
  const domainHex = labels.flatMap(label => {
    const labelHex = Array.from(label).map(c => c.charCodeAt(0));
    return [label.length, ...labelHex];
  });
  domainHex.push(0x00); // End the domain with a zero-length label

  const typeAAAA = new Uint8Array([0x00, 0x1C]); // Type AAAA (IPv6)
  const classIN = new Uint8Array([0x00, 0x01]);  // Class IN
  const ttl = new Uint8Array([0x00, 0x00, 0x00, 0x3C]); // TTL (60 seconds as an example)

  // Convert the string "test" to a 16-byte IPv6 address (padding if necessary)
  const ipv6Bytes = stringToIPv6Bytes("ipv6");

  const rdLength = new Uint8Array([0x00, 0x10]); // Length of IPv6 address (16 bytes)

  return new Uint8Array([
    ...domainHex,
    ...typeAAAA,
    ...classIN,
    ...ttl,
    ...rdLength,
    ...ipv6Bytes,
  ]);
}

// Helper function to convert the string "test" to a 16-byte IPv6 address
function stringToIPv6Bytes(ipv6String) {
  const ipv6Array = Array.from(ipv6String).map(c => c.charCodeAt(0));  // Convert each character to its ASCII code

  // Ensure the result is 16 bytes long by padding with zeros if necessary
  while (ipv6Array.length < 16) {
    ipv6Array.push(0x00);  // Add padding byte
  }

  return new Uint8Array(ipv6Array.slice(0, 16));  // Ensure exactly 16 bytes
}

// Helper function to extract the transaction ID (first two bytes of the DNS query)
function extractTransactionId(queryBuffer) {
  const transactionId = queryBuffer.slice(0, 2);
  return transactionId;
}

// Helper function to extract the domain name from the DNS query body
function extractDomainNameFromQuery(queryBuffer) {
  let offset = 12; // DNS headers are 12 bytes long, so the domain name starts at byte 13
  const domainParts = [];

  while (true) {
    const labelLength = queryBuffer[offset]; // Length of the next label
    if (labelLength === 0) break; // End of domain name (0 byte)

    const label = String.fromCharCode(...queryBuffer.slice(offset + 1, offset + 1 + labelLength));
    domainParts.push(label);
    offset += labelLength + 1; // Move to the next label
  }

  const domain = domainParts.join('.');
  return domain;
}

// Helper function to compare transaction IDs
function compareTransactionId(transactionId, storedTransactionId) {
  if (transactionId.length !== storedTransactionId.length) {
    return false;
  }
  for (let i = 0; i < transactionId.length; i++) {
    if (transactionId[i] !== storedTransactionId[i]) {
      return false;
    }
  }
  return true;
}

// Helper function to encode a domain name into DNS label format
function encodeDomainName(domain) {
  const parts = domain.split('.');
  const labels = [];
  
  for (const part of parts) {
    const length = part.length;
    labels.push(length);  // Push the length of the label
    for (let i = 0; i < length; i++) {
      labels.push(part.charCodeAt(i));  // Push the ASCII value of each character
    }
  }

  labels.push(0x00);  // End the domain name with a null byte
  
  return labels;
}

function parseIPv6(ipv6) {
  const segments = ipv6.split(':');
  const expandedSegments = [];

  let hasEmptySegment = false;
  for (const segment of segments) {
    if (segment === '') {
      if (!hasEmptySegment) {
        hasEmptySegment = true;
        expandedSegments.push(...Array(8 - segments.filter(s => s !== '').length).fill('0'));
      }
    } else {
      expandedSegments.push(segment.padStart(4, '0'));
    }
  }

  return expandedSegments.flatMap(part => {
    const segment = parseInt(part, 16);
    return [(segment >> 8) & 0xff, segment & 0xff];
  });
}
