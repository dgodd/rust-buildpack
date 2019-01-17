use tiny_http::{Server, Response};

fn main() {
    let server = Server::http("0.0.0.0:8080").unwrap();

    for request in server.incoming_requests() {
        println!("received request! : {:?}", request.url());

        let response = Response::from_string("hello world");
        request.respond(response).unwrap();
    }
}
