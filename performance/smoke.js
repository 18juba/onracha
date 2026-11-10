import http from 'k6/http';
import { check, fail } from 'k6';

const configuredBaseURL = __ENV.BASE_URL || 'http://localhost:8080';
const baseURL = configuredBaseURL.endsWith('/')
  ? configuredBaseURL.slice(0, -1)
  : configuredBaseURL;
const password = 'onracha-smoke-password-123';

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: {
    checks: ['rate>0.99'],
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<1000']
  }
};

function postJSON(path, payload) {
  return http.post(baseURL + path, JSON.stringify(payload), {
    headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
    timeout: '5s'
  });
}

function getJSON(path) {
  return http.get(baseURL + path, {
    headers: { Accept: 'application/json' },
    timeout: '5s'
  });
}

function assertCheck(data, name, condition) {
  if (!check(data, { [name]: condition })) {
    fail(name);
  }
}

function expectStatus(response, operation, expectedStatus) {
  assertCheck(
    response,
    operation + ' returns HTTP ' + expectedStatus,
    (result) => result.status === expectedStatus
  );
  return response.json();
}

function registerUser(name, email) {
  const user = expectStatus(
    postJSON('/api/v1/auth/register', { name, email, password }),
    'register ' + name,
    201
  );
  assertCheck(user, name + ' has an id', (result) => result.id > 0);
  return user;
}

function loginUser(name, email) {
  const user = expectStatus(
    postJSON('/api/v1/auth/login', { email, password }),
    'login ' + name,
    200
  );
  assertCheck(user, name + ' session is established', (result) => result.id > 0);
  return user;
}

export default function () {
  const runID = Date.now() + '-' + __VU + '-' + __ITER;
  const organizerName = 'ON Racha CI Organizer';
  const organizerEmail = 'smoke-organizer-' + runID + '@example.test';
  const playerName = 'ON Racha CI Player';
  const playerEmail = 'smoke-player-' + runID + '@example.test';

  const organizer = registerUser(organizerName, organizerEmail);
  loginUser(organizerName, organizerEmail);

  const arena = expectStatus(
    postJSON('/api/v1/arenas', {
      name: 'Arena Smoke ' + runID,
      arena_type: 'society',
      address: 'Rua do Racha, 100',
      city: 'Campinas',
      state: 'SP'
    }),
    'organizer creates an arena',
    201
  );
  assertCheck(arena, 'created arena has an id', (result) => result.id > 0);

  const startsAt = new Date(Date.now() + 60 * 60 * 1000).toISOString();
  const endsAt = new Date(Date.now() + 2 * 60 * 60 * 1000).toISOString();
  const timeSlot = expectStatus(
    postJSON('/api/v1/arenas/' + arena.id + '/time-slots', {
      starts_at: startsAt,
      ends_at: endsAt
    }),
    'organizer creates an available time slot',
    201
  );
  assertCheck(timeSlot, 'created time slot has an id', (result) => result.id > 0);

  const game = expectStatus(
    http.post(baseURL + '/api/v1/time-slots/' + timeSlot.id + '/games', null, { timeout: '5s' }),
    'organizer schedules a game',
    201
  );
  assertCheck(
    game,
    'organizer is automatically enrolled',
    (result) => result.organizer_user_id === organizer.id &&
      result.participant_user_ids.includes(organizer.id)
  );

  const player = registerUser(playerName, playerEmail);
  loginUser(playerName, playerEmail);

  const discoveredGames = expectStatus(
    getJSON('/api/v1/games?city=' + encodeURIComponent(arena.city)),
    'player discovers the game',
    200
  );
  assertCheck(
    discoveredGames,
    'scheduled game appears in search',
    (results) => Array.isArray(results) &&
      results.some((result) => result.game && result.game.id === game.id)
  );

  const joinedGame = expectStatus(
    http.post(baseURL + '/api/v1/games/' + game.id + '/join', null, { timeout: '5s' }),
    'player joins the game',
    200
  );
  assertCheck(
    joinedGame,
    'player appears in the participant list',
    (result) => result.participant_user_ids.includes(player.id)
  );

  const myGames = expectStatus(getJSON('/api/v1/me/games'), 'player lists their games', 200);
  assertCheck(
    myGames,
    'joined game appears in the player schedule',
    (results) => Array.isArray(results) && results.some((result) => result.id === game.id)
  );

  const lobby = expectStatus(
    getJSON('/api/v1/games/' + game.id + '/lobby'),
    'player opens the game lobby',
    200
  );
  assertCheck(
    lobby,
    'lobby includes organizer and player',
    (result) => result.game.id === game.id &&
      result.participants.some((participant) => participant.id === organizer.id) &&
      result.participants.some((participant) => participant.id === player.id)
  );
}
