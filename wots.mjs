/*
Minified version. 
From https://geraintluff.github.io/sha256/
*/
function sha256(r){function h(r,h){return r>>>h|r<<32-h}for(var t,n,o=Math.pow,a=o(2,32),
e=[],f=8*r.length,l=sha256.h=sha256.h||[],g=sha256.k=sha256.k||[],s=g.length,c={},i=2;s<64;i++)
if(!c[i]){for(t=0;t<313;t+=i)c[t]=i;l[s]=o(i,.5)*a|0,g[s++]=o(i,1/3)*a|0}for(r+="";r.length%64-56;)r+="\0";for(t=0;t<r.length;t++){if((n=r.charCodeAt(t))>>8)return;
e[t>>2]|=n<<(3-t)%4*8}for(e[e.length]=f/a|0,e[e.length]=f,n=0;n<e.length;){var u=e.slice(n,n+=16),v=l;for(l=l.slice(0,8),t=0;t<64;t++)
{var k=u[t-15],p=u[t-2],d=l[0],w=l[4],A=l[7]+(h(w,6)^h(w,11)^h(w,25))+(w&l[5]^~w&l[6])+g[t]+(u[t]=t<16?u[t]:u[t-16]+(h(k,7)^h(k,18)^k>>>3)+u[t-7]+(h(p,17)^h(p,19)^p>>>10)|0);
(l=[A+((h(d,2)^h(d,13)^h(d,22))+(d&l[1]^d&l[2]^l[1]&l[2]))|0].concat(l))[4]=l[4]+A|0}for(t=0;t<8;t++)l[t]=l[t]+v[t]|0}var C=[];
for(t=0;t<8;t++)for(n=3;n+1;n--){var M=l[t]>>8*n&255;C.push(M),(M<16?0:"")+M.toString(16),0}return C}

/*
from: https://stackoverflow.com/a/16436975
*/
/*
function arraysEqual(a, b) {
	if (a === b) return true;
	if (a == null || b == null) return false;
	if (a.length !== b.length) return false;
  
	for (var i = 0; i < a.length; ++i) {
	  if (a[i] !== b[i]) return false;
	}
	return true;
}
*/

function bArr_toString(byte_array) {
	var return_str = "";
	for (var i = 0; i < byte_array.length; i++) {
		return_str += String.fromCharCode(byte_array[i]);
	}
	return return_str;
}
//I did also this to make sometimes everything simpler
Array.prototype.toASCII = function () {
	var return_str = "";
	for (var i = 0; i < this.length; i++) {
		return_str += String.fromCharCode(this[i]);
	}
	return return_str; 
}
String.prototype.toBytes = function () {
	var return_bytes = [];
	for (var i = 0; i < this.length; i++) {
		return_bytes.push(this.charCodeAt(i));
	}
	return return_bytes; 
}
Array.prototype.pushArray = function(arr) {
    this.push.apply(this, arr);
};

// NOW THERE ARE THE DEFINITIONS

var PARAMSN = 32;
var WOTSW = 16;
var WOTSLOGW = 4;
var WOTSLEN1 = (8 * PARAMSN / WOTSLOGW);
var WOTSLEN2 = 3;
var WOTSLEN  = WOTSLEN1 + WOTSLEN2;
var WOTSSIGBYTES = WOTSLEN * PARAMSN;

/* 2144 + 32 + 32 = 2208 */
var TXSIGLEN  = 2144;
var TXADDRLEN = 2208;

var XMSS_HASH_PADDING_F = 0;
var XMSS_HASH_PADDING_PRF = 3;


/* the seed is the private key. The seed its 32 bytes of 8 bits each */
function wots_public_key_gen(seed, pub_seed, addr_bytes) {
	var addr = bytes_to_addr(addr_bytes);
    var private_key = expand_seed(seed); // I don't know if this is really the private key. Boh
	var cache_pk = [];
	for (var i = 0; i < WOTSLEN; i++) {
		set_chain_addr(i, addr); //here is not in byte array format cause I think I dont necessarly have to
		var priv_key_portion = private_key.slice(i*PARAMSN, PARAMSN + i*PARAMSN);
		var array_to_push = gen_chain(priv_key_portion, 0, WOTSW - 1, pub_seed, addr );

		cache_pk.pushArray(array_to_push);
	}
	return cache_pk;
}

function wots_sign(msg, seed, pub_seed, addr_bytes) {
	var addr = bytes_to_addr(addr_bytes);
	var lenghts = []; //its WOTSLEN long (67)
	var signature = [];
	lenghts = chain_lenghts(msg);
	console.log("lenghts");
	console.log(lenghts.length);
	console.log(lenghts.toString());
	/* the wots private key comes from the seed*/
	var private_key = expand_seed(seed); // this is the private key

	for(var i = 0; i < WOTSLEN; i++) {
		set_chain_addr(i, addr);
		var priv_key_portion = private_key.slice(i*PARAMSN, PARAMSN + i*PARAMSN);
		var array_to_push = gen_chain(priv_key_portion, 0, lenghts[i], pub_seed, addr );
		signature.pushArray(array_to_push);
	}
	return signature;
}

function wots_publickey_from_sig(sig, msg, pub_seed, addr_bytes) {
	var addr = bytes_to_addr(addr_bytes);
	var lenghts = []; //array of WOTSLEN size
	lenghts = chain_lenghts(msg);

	var public_key = [];
	for(var i = 0; i < WOTSLEN; i++) {
		set_chain_addr(i, addr);
		var signature_portion = sig.slice(i*PARAMSN, PARAMSN + i*PARAMSN);
		public_key.pushArray(gen_chain(signature_portion, lenghts[i], WOTSW - 1 - lenghts[i], pub_seed, addr));
	}
	return public_key;
}

function expand_seed(seed) {
    var ctr = []; //This will be max 32 items
    var out_seeds = []; //This will be maximum with WOTSLEN items
    for(var i = 0; i < WOTSLEN; i++) {
        ctr = ull_to_bytes(PARAMSN, [i]); //yeah I hope "i" doesnt go more than 255
		out_seeds.pushArray(prf(ctr, seed));
    }
	return out_seeds;
}

function ull_to_bytes(outlen, input) {
    var out_array = [];
	for(var i = outlen-1; i>=0; i--) {
		var to_push = input[i];
		if(to_push == undefined) {
			out_array.push(0);
		} else {
			out_array.push(to_push);
		}
	}
    return out_array;
}

function prf(input, key) {
    var buf = [];
    buf = ull_to_bytes(PARAMSN, [XMSS_HASH_PADDING_PRF]); // 32 and 3

	var byte_copied_key = byte_copy(key, PARAMSN);
	buf.pushArray(byte_copied_key);

	var byte_copied_input = byte_copy(input,32);
	buf.pushArray(byte_copied_input);

	//I was dubios of this but it works as expected
	return sha256(buf.toASCII());
}
function t_hash(input, pub_seed, addr) {
	var buf = []; // maximuym lenght 3*PARAMSN
	var bitmask = []; //maximum lenght PARAMSN
	var addr_as_bytes = []; //maximum lenght 32

	var buf = ull_to_bytes(PARAMSN, [XMSS_HASH_PADDING_F]);

	/* generate n-byte key */
	set_key_and_mask(0, addr);
	addr_as_bytes = addr_to_bytes(addr);
	var to_push_buf = prf(addr_as_bytes, pub_seed);
	buf.pushArray(to_push_buf);

	/*generate the n-byte mask */
	set_key_and_mask(1, addr);
	addr_as_bytes = addr_to_bytes(addr);
	var bitmask = prf(addr_as_bytes, pub_seed);

	var XOR_bitmask_input = []; 
	for(var i = 0; i < PARAMSN; i++) {
		XOR_bitmask_input.push(input[i] ^ bitmask[i]);
	}
	buf.pushArray(XOR_bitmask_input);
	return sha256(buf.toASCII());
}

function byte_copy(source, num_bytes) {
    var output = []
    for(var i = 0; i < num_bytes; i++) {
		if(source[i] == undefined) {
			output.push(0);
		} else{
			output.push(source[i]);
		}
    }
	return output;
}

function gen_chain(input, start, steps, pub_seed, addr)  {
	var out = byte_copy(input, PARAMSN);
	
	for ( var i = start; i < (start+steps) && i < WOTSW; i++) {
		set_hash_addr(i, addr);
		out = t_hash(out, pub_seed, addr);
	}
	return out;
}

/*
This works only if log_w is a divisor of 8!
*/
function base_w(outlen, input) {
	/*
	if(8 % WOTSLOGW != 0) {
		console.log("SECURITY PROBLEM! WOTSLOGW MUST BE A DIVISOR OF 8!");
	}
	thinking about then making an upper library that handles all those functions with objects, so I check that automatically
	*/
	var in_ = 0;
	var out = 0;
	var total; // I am not sure about this one
	var bits = 0;
	var output = [];
	for(var consumed = 0; consumed < outlen; consumed++) {
		if(bits == 0) {
			total = input[in_];
			in_ ++;
			bits += 8;
		}
		bits -= WOTSLOGW;
		output[out] = (total >> bits) & (WOTSW - 1);
		out++;
	}
	return output;
}
function wots_checksum(msg_base_w) {
	var csum = 0;
	var csum_bytes = []; //array of size (WOTSLEN2 * WOTSLOGW + 7) / 8 = (12 + 7) / 8 --> mhhhhhhh
	for (var i = 0; i < WOTSLEN1; i++) {
		csum += WOTSW - 1 - msg_base_w[i];
	}
	/* convert checksum to base_w */
	csum = csum << (8 - ((WOTSLEN2 * WOTSLOGW) % 8));
	//console.log(csum);
	csum_bytes = ull_to_bytes(Math.round((WOTSLEN2 * WOTSLOGW + 7) / 8), from_int_to_byte_array(csum)); 
	// ^^^ TO CHECK! from_int_to_byte_array makes the most significant byte in the end!
	
	//console.log("csum bytes: " + csum_bytes.toString());
	var csum_base_w = base_w(WOTSLEN2, csum_bytes);
	return csum_base_w;
}
function chain_lenghts(msg) {
	var lenghts = base_w(WOTSLEN1, msg);
	lenghts.pushArray(wots_checksum(lenghts));
	return lenghts;
}

function set_chain_addr(chain_address, addr) { 
	addr["5"] = [0,0,0,chain_address]; //yeah hard coded but its ok bro
}
function set_hash_addr(hash, addr) {
	addr["6"] = [0,0,0,hash];
}
function set_key_and_mask(key_and_mask, addr) {
	addr["7"] = [0,0,0,key_and_mask];
}
function from_int_to_byte_array(number) {
	var out_array = [];
	if(number==0) {
		out_array.push(number);
	}
	while(number != 0) {
		out_array.push(number & 0xff);
		number = number >> 8;
	}
	return out_array;
}
function addr_to_bytes(addr) {
	var out_bytes = [];
	for(var i = 0; i < 8; i++) {
		if(addr[(i.toString())] == undefined) {
			addr[i.toString()] = [0,0,0,0];
		}
		var to_push = addr[(i.toString())];
		out_bytes.pushArray(to_push);
	}
	return out_bytes;
}
function bytes_to_addr(addr_bytes) {
	var out_addr = {"0":[0,0,0,0], "1":[0,0,0,0], "2":[0,0,0,0], "3":[0,0,0,0], 
					"4":[0,0,0,0], "5":[0,0,0,0], "6":[0,0,0,0], "7":[0,0,0,0]};
	for(var i = 0; i < 8; i++) {
		out_addr[i.toString()] =  ull_to_bytes(4, addr_bytes.slice(i*4, i*4 + 4));
	}
	return out_addr;
}

export {sha256, wots_sign, wots_public_key_gen, wots_publickey_from_sig, ull_to_bytes, from_int_to_byte_array, byte_copy};