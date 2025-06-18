document.addEventListener('DOMContentLoaded', () => {
    const movieList = document.getElementById('movie-list');

    fetch('/movies/')
        .then(response => response.json())
        .then(movies => {
            if (movies.length === 0) {
                movieList.innerHTML = '<p>No movies available at the moment.</p>';
                return;
            }

            const ul = document.createElement('ul');
            movies.forEach(movie => {
                const li = document.createElement('li');
                li.innerHTML = `<strong>${movie.title}</strong> (<em>${movie.genre}</em>) - Directed by ${movie.director}`;
                ul.appendChild(li);
            });
            movieList.appendChild(ul);
        })
        .catch(error => {
            console.error('Error fetching movies:', error);
            movieList.innerHTML = '<p>Could not load movies. Please try again later.</p>';
        });
});
