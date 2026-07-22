import * as http from 'https';

function login(email: string): Promise<string> {
  return new Promise((resolve, reject) => {
    const data = JSON.stringify({
      email,
      password: 'ProdTest123!',
    });

    const req = http.request(
      'https://gkcircle.com/api/v1/auth/login',
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Content-Length': data.length,
        },
      },
      (res) => {
        let body = '';
        res.on('data', (chunk) => (body += chunk));
        res.on('end', () => {
          if (res.statusCode === 201 || res.statusCode === 200) {
            const parsed = JSON.parse(body);
            console.log('API Response Roles:', parsed.data?.user?.roles);
            resolve(parsed.data?.token || parsed.token || parsed.accessToken || parsed.data?.accessToken);
          } else {
            reject(new Error(`Login failed: ${res.statusCode} ${body}`));
          }
        });
      }
    );

    req.on('error', reject);
    req.write(data);
    req.end();
  });
}

function parseJwt(token: string) {
  return JSON.parse(Buffer.from(token.split('.')[1], 'base64').toString());
}

async function main() {
  const emailA = 'creator-a-mq7a6bf0@prod.gkcircle.com';
  const emailB = 'creator-b-mq7a6bf0@prod.gkcircle.com';

  const tokenA = await login(emailA);
  const tokenB = await login(emailB);

  console.log('Creator A JWT Payload:');
  console.log(JSON.stringify(parseJwt(tokenA), null, 2));

  console.log('\nCreator B JWT Payload:');
  console.log(JSON.stringify(parseJwt(tokenB), null, 2));
}

main().catch(console.error);
