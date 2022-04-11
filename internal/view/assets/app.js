new Vue({
    el: '#app',
    data() {
        return {
            input: null,
            order: null,
            loading: false,
            errored: false
        };
    },
    methods: {
        update: _.debounce(function (e) {
            this.input = e.target.value;
            this.getOrder(this.input)
        },500),
        getOrder(id) {
            axios.get(`/order/${id}`)
                .then(response => {
                    this.errored = false;
                    this.order = JSON.parse(response.data);
                })
                .catch(error => {
                    this.errored = true;
                    this.order = false
                })
                .finally(() => (
                    this.loading = true
                ));
        }
    }
});