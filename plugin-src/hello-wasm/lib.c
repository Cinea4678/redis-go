#include "tinyalloc.h"

extern void set_key(const char* key, const char* value);
extern void set_key_int(const char* key, long long value);
extern void delete_key(const char* key);
extern const char* get_key(const char* key);
extern long long get_key_int(char key[]);
extern void set_expire(const char* key, int expire);
extern int get_expire(const char* key);

// C的wasm开发环境比较简陋，不能使用stdio等系统库，因此我们手动实现一个超级简陋的itoa
void simple_itoa(char* buffer, int val) {
	int nums[3];
	nums[0] = (val/100) % 10;
	nums[1] = (val/10) % 10;
	nums[2] = val % 10;

	for(int i=0;i<3;i++){
		if(nums[i]==0 && i!=2) {
			buffer[i] = ' ';
		} else {
			buffer[i] = '0'+nums[i];
		}
	}
}

extern unsigned char MEMORY[12 * 1024];
static int ALLOC_INITED;

void* alloc(int length) {
	if(ALLOC_INITED==0){
		ALLOC_INITED=1;
		ta_init(MEMORY, MEMORY+12*1024, 256, 1024, 4);
	}
	return ta_alloc(length);
}

void free(void* ptr) {
	ta_free(ptr);
}

const char* info() {
	return "{\"name\":\"hello-c\", \"commands\":[\"hellowasm\"]}";
}

const char* handle(const char* req) {
	int count = get_key_int("wasm_count");
	if(count==0){
		set_key_int("wasm_count", 1);
		return "\r\n+It's your first time to use me :)\r\n";
	} else {
		char* buffer = "\r\n+You have called this command     times.\r\n";
		/* sprintf(buffer, "\r\n+You have called this command %d times.\r\n", count); */
		simple_itoa(buffer+32, count);
		set_key_int("wasm_count", count+1);
		return buffer;
	}
}

