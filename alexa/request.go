package alexa

// RequestEnvelope is the deserialized http post request sent by alexa.
type RequestEnvelope struct {
	Version string  `json:"version"`
	Session Session `json:"session"`
	// one of the request structs
	Request Request `json:"request"`
	Context Context `json:"context"`
}

// Session object contained in standard request types like LaunchRequest, IntentRequest, SessionEndedRequest and GameEngine interface.
type Session struct {
	New         bool                   `json:"new"`
	SessionID   string                 `json:"sessionId"`
	Attributes  map[string]interface{} `json:"attributes"`
	Application Application            `json:"application"`
	User        User                   `json:"user"`
}

// Application object with the applications unique id.
type Application struct {
	ApplicationID string `json:"applicationId"`
}

// User contains the userId and access token if existent.
type User struct {
	UserID      string `json:"userId"`
	AccessToken string `json:"accessToken,omitempty"`
}

// Context object provides your skill with information about the current state of the Alexa service and device at the time the request is sent to your service.
type Context struct {
	System      System      `json:"System"`
	AudioPlayer AudioPlayer `json:"audioPlayer"`
}

// System object that provides information about the current state of the Alexa service and the device interacting with your skill.
type System struct {
	APIAccessToken string      `json:"apiAccessToken"`
	APIEndpoint    string      `json:"apiEndpoint"`
	Application    Application `json:"application"`
	Device         Device      `json:"device"`
	User           User        `json:"user"`
}

// Device object providing information about the device used to send the request.
type Device struct {
	DeviceID            string                 `json:"deviceId"`
	SupportedInterfaces map[string]interface{} `json:"supportedInterfaces"`
}

// AudioPlayer object providing the current state for the AudioPlayer interface.
type AudioPlayer struct {
	Token                string `json:"token,omitempty"`
	OffsetInMilliseconds int    `json:"offsetInMilliseconds,omitempty"`
	PlayerActivity       string `json:"playerActivity"`
}

// Request contains the attributes all alexa requests have in common.
type Request struct {
	Type      string `json:"type"`
	RequestID string `json:"requestId"`
	Timestamp string `json:"timestamp"`
	Locale    string `json:"locale"`
	// Set manually from request envelope
	Session *Session `json:"session,omitempty"`
	Context *Context `json:"context,omitempty"`
	// Intent Requests
	Intent      Intent `json:"intent,omitempty"`
	DialogState string `json:"dialogState,omitempty"`
	// SessionEndRequest
	Reason string `json:"reason,omitempty"`
	// SessionEndRequest, SystemExceptionEncounteredRequest
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
	// SystemExceptionEncounteredRequest
	Cause struct {
		RequestID string `json:"requestId"`
	} `json:"cause"`

	// AudioPlayerRequest represents an incoming request from the Audioplayer Interface.
	// It does not have a session context.  Response to such a request must be a
	// AudioPlayerDirective or empty
	Token                string `json:"token"`
	OffsetInMilliseconds int    `json:"offsetInMilliseconds"`

	// AudioPlayerPlaybackFailedRequest is sent when Alexa encounters an error when attempting to play a stream.
	CurrentPlaybackState struct {
		Token                string `json:"token"`
		OffsetInMilliseconds int    `json:"offsetInMilliseconds"`
		PlayerActivity       string `json:"playerActivity"`
	} `json:"currentPlaybackState"`
}

// Intent provided in Intent requests
type Intent struct {
	Name               string                `json:"name,omitempty"`
	Slots              map[string]IntentSlot `json:"slots,omitempty"`
	ConfirmationStatus string                `json:"confirmationStatus,omitempty"`
}

// IntentSlot is provided in Intents
type IntentSlot struct {
	Name               string      `json:"name"`
	Value              string      `json:"value"`
	ConfirmationStatus string      `json:"confirmationStatus,omitempty"`
	Resolutions        interface{} `json:"resolutions"`
}
