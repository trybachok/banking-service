package cbr

func keyRateSOAPEnvelope() string {
	return `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
               xmlns:xsd="http://www.w3.org/2001/XMLSchema"
               xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <KeyRate xmlns="http://web.cbr.ru/">
      <fromDate>2024-01-01T00:00:00</fromDate>
      <ToDate>2030-01-01T00:00:00</ToDate>
    </KeyRate>
  </soap:Body>
</soap:Envelope>`
}
