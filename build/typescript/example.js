var example;
(function (example) {
    function init() {
        example.word = "bird is the word.";
        example.HaveYouNotHeard();
    }
    example.init = init;
    function HaveYouNotHeard() {
        console.log("Have you not heard that " + example.word);
    }
    example.HaveYouNotHeard = HaveYouNotHeard;
})(example || (example = {}));
