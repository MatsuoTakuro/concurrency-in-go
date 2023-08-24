package main

// sendEmail sends a message to the Mailer's channel
func (s *Server) sendEmail(msg Message) {
	if !s.Mailer.canAcceptMessage() {
		s.InfoLog.Println("New message not accepted")
		return
	}
	s.Wait.Add(1) // increment counter every time a new message is sent
	s.Mailer.Msg <- msg
}
