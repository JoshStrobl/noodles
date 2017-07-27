// This is just an example Typescript file to test compilation

namespace example {
	export var word: string;

	// Initialize
	export function init() {
		example.word = "bird is the word.";
		example.HaveYouNotHeard();
	}

	// HaveYouNotHeard will tell people the word
	export function HaveYouNotHeard() {
		console.log(`Have you not heard that ${example.word}`);
	}
}