import { PostsApi, Configuration } from '../client/index';

const AUTH_TOKEN = 'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IjUwTDFCYzNiZ0U3dng1TDRmSTdHTyJ9.eyJpc3MiOiJodHRwczovL2ZvcmNlLXN0cmVhbS51cy5hdXRoMC5jb20vIiwic3ViIjoidHdpdHRlcnwxMzgyOTQzMTg1MzQxNTk5NzQ2IiwiYXVkIjpbImh0dHBzOi8vcHJvZHVjZS50aHJlYWRtaXJyb3IuY29tIiwiaHR0cHM6Ly9mb3JjZS1zdHJlYW0udXMuYXV0aDAuY29tL3VzZXJpbmZvIl0sImlhdCI6MTc1MDkyNzYwOCwiZXhwIjoxNzUxMDE0MDA4LCJzY29wZSI6Im9wZW5pZCBwcm9maWxlIGVtYWlsIiwiYXpwIjoiTG9BV3dSWGp1eVJrU0lxbFkydTExalRYaml5RnhZd2wifQ.hVHQHJmnOCfsDKlC5tS-8xKwbz7K8zJlSyuTlc0LYodK6VBaW3ACbgICgi9Iy7vPcfLh9knoBtIchjHKJ9CR5LgT20Ul2l8MSRjTy6zo29oNEIEJbOPwO04u1AgxTzCxyFYrWL3_XD59A-1JwWci-Q8720tomULz0JMnZWGCOCUlKAiaAgZrchmskXo-BBhi3Nwn42Mn6HUJikBunjRYLUm8Y5vP_rKeZStIfErWAwD0XSkqwoEF3yDV1tfzIdtAJKDip9oRxNtN2LBJdkGM4xnn_kUE_lAg5IYI3wLGi2jsWpuYBPxrd50LHCV5Y_MNIXVu6QbLYwg3ppPHFJRgxg';

const config = new Configuration({
    basePath: 'http://localhost:8080/api/v1',
    accessToken: AUTH_TOKEN,
    middleware: [
        {
            pre: async (context) => {
                // get token from cookie
                const token = await config.accessToken?.();
                context.init.headers = {
                    ...context.init.headers,
                    'Authorization': `Bearer ${token}`
                };
            }
        }
    ]
});

const postApi = new PostsApi(config);

const postApiTest = async () => {
    const response = await postApi.postsGet({
        limit: 10,
        offset: 0,
    });
    console.log(response);
};

(async () => {
    await postApiTest();
})();