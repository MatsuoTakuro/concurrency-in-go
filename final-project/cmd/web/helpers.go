package main

func (s *Server) sendEmail(msg Message) {
	if !s.Mailer.canAcceptMessage() {
		s.InfoLog.Println("New message not accepted")
		return
	}
	s.SentMail.Add(1)
	s.Mailer.MsgChan <- msg
}
